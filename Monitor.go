package main

import( 
	"log"
	"github.com/zeromq/goczmq"
	"encoding/json"
	"strings"
	//"strconv"	
	"sync"
	"time"
)

type Monitor struct {
	mu sync.Mutex
	cond *sync.Cond
	sender *goczmq.Sock 
	broadcaster *goczmq.Sock
	receiver *goczmq.Sock
	id string
	addr string
	//address [3]string
	address map[string]string 
	RN map[string]int 
	sn int `default:0`
	haveToken bool `default:false`
	token Token `default:Token{}`
	asked bool `default:false`
	blocked bool `default:false`
	waitedThreads int `default:0`
}


func (m *Monitor) initMonitor(){
	m.mu = sync.Mutex{}
	m.cond = sync.NewCond(&m.mu)

	destination := ""
	
	for _,i := range m.address{
		if i != m.addr{
			destination += "tcp://" + i + ","
		}
	}
	/*
	for i := 0; i < len(m.address); i++ {
		if m.address[i] != m.addr{
			destination += "tcp://" + m.address[i] + ","
		}        
    }*/
	destination = strings.TrimRight(destination, ",")


	sources := "tcp://"+m.addr

	/*Create ZMQ network socket*/
	
	s, err := goczmq.NewPush(destination)
	if err != nil {
		log.Fatal(err)
	}
	m.sender = s

	b, err := goczmq.NewPush(destination)
	if err != nil {
		log.Fatal(err)
	}
	m.broadcaster = b

	r, err := goczmq.NewPull(sources)
	if err != nil {
		log.Fatal(err)
	}
	m.receiver = r
	go m.listenRequests()
	time.Sleep(time.Millisecond * 100)
	go m.periodicNotify()	
}

func (m *Monitor) DestroyMonitor(){
	
	m.broadcaster.Destroy()
	m.sender.Destroy()
	m.receiver.Destroy()

}

func (m* Monitor)periodicNotify(){
	for {
		time.Sleep(time.Millisecond*20)
		m.cond.Signal()
	}
}

func (m *Monitor) askAboutToken() {

	if !m.asked && !m.haveToken{
		log.Printf("'%s': Ask about token",m.id)
		m.sendRequest()
		m.asked = true
	}
}

func (m *Monitor) lock(){
	m.cond.L.Lock()
	m.waitedThreads+=1
}

func (m *Monitor) unlock(){
	m.waitedThreads-=1
	m.cond.L.Unlock()
}

func (m *Monitor) wait(){
	m.blocked = true
	m.askAboutToken()
	m.cond.Wait()
}

func (m *Monitor) notifyAll(){
	m.blocked = false
	m.cond.Broadcast()
	
}

func (m *Monitor) notify(){
	m.blocked = false
	m.cond.Signal()
}

func (m *Monitor) sendToken(){

	var id string = m.token.TopToken()
	//i, _ := strconv.Atoi(id)
	destination := "tcp://" + m.address[ id ]
	
	m.sender.Destroy()

	dealer, err := goczmq.NewPush(destination)
	if err != nil {
		log.Fatal(err)
	}
	m.sender = dealer
	
	stringToken := m.token.TokenToString()

	data := make(map[string]interface{})
	data = map[string]interface{}{
        "id": m.id,
		"type":"Token",
		"sn": 0,
		"token":stringToken,
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
        log.Fatal(err)
    }

	err = m.sender.SendFrame([]byte(jsonBytes), goczmq.FlagNone)
	if err != nil {
		log.Fatal(err)
	}
	m.asked = false

}

func (m *Monitor) sendRequest(){
	
	/* Formating ip adresses */
	/*
	destination := ""
	for i := 0; i < len(m.address); i++ {
		if m.address[i] != m.addr{
			destination += "tcp://" + m.address[i] + ","
		}
        
    }
	destination = strings.TrimRight(destination, ",")
	*/
	/*Increment Sequence Number*/
	m.sn = int(m.sn + 1)
	m.RN[m.id]=m.sn

	/*Serialize data before send*/
	data := make(map[string]interface{})
	data = map[string]interface{}{
        "id": m.id,
		"type":"Request",
		"sn": int(m.sn),
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
        log.Fatal(err)
    }
	log.Printf("'%s': Send Request",m.id)
	/*Send data*/
	for i:=1; i<len(m.address); i++ {
		err = m.sender.SendFrame([]byte(jsonBytes), goczmq.FlagNone)
		if err != nil {
			log.Fatal(err)
		}
	//log.Println(string(m.id),"sent",string(jsonBytes),"to",destination)	
	}

	
}

func (m *Monitor) CS(){

	m.mu.Lock()
	log.Printf("'%s': I enter in Critical Section",m.id)
	
	m.haveToken = true
	m.blocked = false
	m.cond.Broadcast()
	
	m.mu.Unlock()

	//Time for operations
	for !m.blocked && m.waitedThreads!=0{
		time.Sleep(time.Millisecond*20)
	}

	m.mu.Lock()

	log.Printf("'%s': I EXIT Critical Section",m.id)
	//log.Println(m.RN,m.token.LN)
		
	for k,v := range m.token.LN{
		if (m.RN[k] == v+1 ){
			if(!m.token.is(k)){
				m.token.EnqueueToken(k)
			}			
		}
	}
	
	m.token.LN[m.id]=m.RN[m.id]
	m.token.DequeueToken()
		
	if !m.token.EmptyToken() {
		m.haveToken = false
		log.Printf("'%s':\033[35m Send Token to '%s', Token: %s \033[0m",m.id,m.token.TopToken(),m.token.TokenToString())
		m.sendToken()
	}else{
		log.Printf("Token queue is empty and wait for requests")
	}
	m.mu.Unlock()
}

func (m *Monitor) listenRequests(){
  
	for{

		log.Printf("'%s': Wait for request",m.id)
		request, err := m.receiver.RecvMessage()
		
		if err != nil {
		  log.Fatal(err)
		}
	
		/*Deserialize data*/
		var p map[string]interface{}
		err = json.Unmarshal(request[0], &p)
		if err!= nil {
			log.Fatal(err)
		}

		switch p["type"]{
			
			case "Request": {
			
				var sn int = int(p["sn"].(float64))
				var id string = p["id"].(string)
			
				log.Println(m.id,m.RN)		
				if ( m.RN[id] < sn  ){
					log.Println(m.id,": Request from",id,"Accept!")
					m.RN[id]=sn

					if (m.haveToken){
						m.mu.Lock()
						
						if (m.haveToken && m.token.EmptyToken()){  // 
							m.haveToken = false
							m.token.EnqueueToken(id)
							//log.Println(m.id,": SEND TOKEN to",id)
							log.Printf("'%s':\033[35m SEND TOKEN to '%s' Token: '%s' \033[0m", m.id, id, m.token.TokenToString())
							m.sendToken()
						}

						m.mu.Unlock()
		
					}else {log.Println(m.id,"I don't have token")}	
				}
			}


			case "Token": {
				var token string = p["token"].(string)
				log.Printf("'%s' HAVE TOKEN NOW from '%s' token='%s'",m.id, p["id"].(string), token)
				
				//log.Printf("'%d'",len(token))
				m.token.stringToToken(token)
				
				go m.CS() 
				
			}
		}
		
	}

}


