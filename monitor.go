package main

import( 
	"log"
	"github.com/zeromq/goczmq"
	"time"
	"encoding/json"
	"strings"
	"strconv"	
	"sync"
)

type Monitor struct {
	mu sync.Mutex
	id string
	addr string
	address [3]string 
	RN map[string]int 
	sn int `default:0`
	haveToken bool `default:false`
	token Token `default:Token{}`
	
}


func (m *Monitor) wait(){

}

func (m *Monitor) sendToken(){

	var id string = m.token.TopToken()
	i, _ := strconv.Atoi(id)
	destination := "tcp://" + m.address[ i-1 ]
	
	dealer, err := goczmq.NewPush(destination)
	if err != nil {
		log.Fatal(err)
	}
	defer dealer.Destroy()
	
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

	err = dealer.SendFrame([]byte(jsonBytes), goczmq.FlagNone)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(10000 * time.Millisecond)
}

func (m *Monitor) sendRequest(){
	
	/* Formating ip adresses */
	destination := ""
	for i := 0; i < len(m.address); i++ {
		if m.address[i] != m.addr{
			destination += "tcp://" + m.address[i] + ","
		}
        
    }
	destination = strings.TrimRight(destination, ",")

	/*Create ZMQ network socket*/
	dealer, err := goczmq.NewPush(destination)
	if err != nil {
		log.Fatal(err)
	}
	defer dealer.Destroy()
	
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

	/*Send data*/
	for i:=1; i<len(m.address); i++ {
		err = dealer.SendFrame([]byte(jsonBytes), goczmq.FlagNone)
		if err != nil {
			log.Fatal(err)
		}
	log.Println(string(m.id),"sent",string(jsonBytes),"to",destination)	
	}

	//log.Println(string(m.id),"sent",string(jsonBytes),"to",destination)
	
	/*Wait and not close connection*/
	time.Sleep(10000 * time.Millisecond)

	/*TO DO : wait for acknowledgement instand time out*/
}

func (m *Monitor) CS(){


	m.mu.Lock()
	m.haveToken = true
	log.Printf("'%s': I enter in Critical Section",m.id)
	soLongOperation()
	log.Printf("'%s': I EXIT Critical Section",m.id)
	m.mu.Unlock()
	
	log.Println(m.RN,m.token.LN)
	
	m.mu.Lock()
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
		log.Printf("'%s': Send Token '%s'",m.id,)
		m.sendToken()
	}
	m.mu.Unlock()

}

func (m *Monitor) listenRequests(){

	/*Formating ip adress*/
	sources := "tcp://"+m.addr

	/*Create ZMQ network socket*/
	router, err := goczmq.NewPull(sources)
	  if err != nil {
		  log.Fatal(err)
	  }
	  defer router.Destroy()
	  
	for{

		log.Printf("'%s': Wait for request",m.id)
		request, err := router.RecvMessage()
		
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
						if (m.haveToken){
							m.haveToken = false
							m.token.EnqueueToken(id)
							log.Println(m.id,": SEND token to",id)
							go m.sendToken()
						}
						m.mu.Unlock()

						
						
						
					}else {log.Println(m.id,"I don't have token")}
					
				}
			} 

			case "Token": {
				
				var token string = p["token"].(string)
				log.Printf("'%s' Got token message from '%s' token='%s'",m.id, p["id"].(string), token)
				m.token.stringToToken(token)
				
				go m.CS() 
				
			}
		}
		
	}

}

type ProducerConsumer struct {
	counter int
	Monitor
	buffor[5] int
	bufforFull bool
	maxSize int `default:5`
}

func (s *ProducerConsumer) produce(product int){
	
	if s.bufforFull{
		//wait
	}
	if !s.bufforFull{
		s.buffor = append(s.buffor, product)
	}
	
}

func (s *ProducerConsumer) consume(product int){
	
	if s.bufforEmpty{
		//s.wait()
	}

	if !s.bufforEmpty{
		temp := s.buffor[0]
    	s.buffor = s.buffor[1:]
	}

}

func (q *ProducerConsumer) bufforEmpty() bool {
    return len(q.queue) == 0
}

func (q *ProducerConsumer) bufforFull() bool {
    return len(q.queue) == q.maxSize
}


func main() {

	// Create a new server
	adresses := [3]string{"127.0.0.1:5555","127.0.0.1:5556","127.0.0.1:5557"}
	
	var RN = map[string]int {
        "1": 0,
        "2": 0, 
		"3": 0,
	}
	RN["1"]=1
	var LN = map[string]int {
        "1": 0,
        "2": 0, 
		"3": 0,
	}
	LN["1"]=1
	products := [5]int{0,0,0,0,0}
	
	token := Token{ make([]string, 0), map[string]int {"1": 0,"2": 0, "3": 0,}}

	producer :=   &ProducerConsumer{0, Monitor{id:"1",addr:"127.0.0.1:5555",address: adresses,RN: map[string]int {"1": 0,"2": 0, "3": 0,},haveToken: true,token: token },products , false}
	consument1 := &ProducerConsumer{0, Monitor{id:"2",addr:"127.0.0.1:5556",address: adresses,RN: map[string]int {"1": 0,"2": 0, "3": 0,},token: token },products , false}
	consument2 := &ProducerConsumer{0, Monitor{id:"3",addr:"127.0.0.1:5557",address: adresses,RN: map[string]int {"1": 0,"2": 0, "3": 0,},token: token },products , false}
	
	//producer.id=""
	//consument1.id=""
	//consument2.id=""


	go producer.listenRequests()
	go consument1.listenRequests()
	go consument2.listenRequests()
	log.Println(producer.RN,consument1.RN,consument2.RN)
	go consument1.sendRequest()
	time.Sleep(1000 * time.Millisecond)
	go consument2.sendRequest()
	time.Sleep(1000 * time.Millisecond)
	go producer.sendRequest()
	go consument1.sendRequest()
	time.Sleep(1000 * time.Millisecond)
	go consument2.sendRequest()
	time.Sleep(1000 * time.Millisecond)
	go producer.sendRequest()
	
//	wg := sync.WaitGroup{}
//	wg.Add(10)

	
	
	time.Sleep(500 * time.Millisecond)
	
//	wg.Wait()

	//go server.send("Hej monitorek","tcp://127.0.0.1:5555")
	//go send("Test 1","tcp://127.0.0.1:5555,tcp://127.0.0.1:5556")
//	go send("Test 2","tcp://127.0.0.1:7777,tcp://127.0.0.1:5555")
	//go send("Test 3","tcp://127.0.0.1:5557")
//	go get("tcp://127.0.0.1:7777")
//	go get("tcp://127.0.0.1:5555")

	time.Sleep(20000 * time.Millisecond)
	log.Println(producer.RN,consument1.RN,consument2.RN)

}

func soLongOperation(){
	time.Sleep(100 * time.Millisecond)
	log.Printf("operation finished successfully")
				
}





func send( msg string, address string){
	dealer, err := goczmq.NewPush(address)
	if err != nil {
		log.Fatal(err)
	}
	defer dealer.Destroy()

  	err = dealer.SendFrame([]byte(msg), goczmq.FlagNone)
	if err != nil {
		log.Fatal(err)
	}
	err = dealer.SendFrame([]byte(msg), goczmq.FlagNone)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("dealer sent '%s' to '%s'",msg, address)

	time.Sleep(10000 * time.Millisecond)
	
  	// Receive the reply.
	/*reply, err := dealer.RecvMessage()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("dealer received '%s'", string(reply[0]))
*/
}

func get(address string) {
	router, err := goczmq.NewPull(address)
	  if err != nil {
		  log.Printf("błąd gniazda")
		  log.Fatal(err)
	  }
	  defer router.Destroy()
	  for{
		log.Println("wait for message")	
		request, err := router.RecvMessage()
	  if err != nil {
		  log.Printf("błąd odbioru")
		  log.Fatal(err)
	  }
	  
	  log.Printf("router received '%s' from '%v'", request[0],address)

	  /* 	err = router.SendFrame(request[0], goczmq.FlagMore)
	  if err != nil {
		  log.Fatal(err)
	  }
  
	  log.Printf("router sent 'World'")
  
	err = router.SendFrame([]byte("World"), goczmq.FlagNone)
	  if err != nil {
		  log.Fatal(err)
	  }
*/
	  }
	
  }

type Token struct {
    queue []string `default:make([]string, 0)`
	LN map[string]int `default:{"1": 0,"2": 0, "3": 0,}`
}


func (q *Token) TokenToString() string{
	result := ""
	for k, v := range q.LN{
		result = result + k + "--" + strconv.Itoa(v) + "$"
	}

	result += "Q:"
	for _, i := range q.queue {
		result = result + string(i) + "$"
	}

	return result
}


func (q *Token) stringToToken(str string){
//"1--0$2--5$3--0$Q:1$2$"

	q.LN = make(map[string]int,0)
	q.queue = make([]string,0)
	endOfLN := strings.LastIndex(str, "Q:")
	
	first_part := strings.Split(str[0:endOfLN-1], "$")

	for _,v := range first_part{
		id_and_lvalue := strings.Split(v, "--")
		q.LN[id_and_lvalue[0]], _ = strconv.Atoi(id_and_lvalue[1])
	}

	second_part := strings.Split(str[endOfLN+2:len(str)-1], "$")

	for _, i := range second_part{
		q.EnqueueToken(i)
	}
}



func (q *Token) is(s string) bool {
    for _,v := range q.queue{
		if v == s {
			return true
		}
	}
	return false
}



func (q *Token) TopToken() string {
    return q.queue[0]
}

func (q *Token) EnqueueToken(s string) {
    q.queue = append(q.queue, s)
}

func (q *Token) DequeueToken() string {
    temp := q.queue[0]
    q.queue = q.queue[1:]
    return temp
}

func (q *Token) EmptyToken() bool {
    return len(q.queue) == 0
}

