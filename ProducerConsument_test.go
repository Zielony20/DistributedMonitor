package main

import "testing"
//import "time"
import "log"
import "sync"

func TestProducerConsument(t *testing.T) {

	//adresses := [3]string{"127.0.0.1:5555","127.0.0.1:5556","127.0.0.1:5557"}
	adresses := map[string]string {"Tomek": "127.0.0.1:5555","Adam": "127.0.0.1:5556", "Krzychu": "127.0.0.1:5557",}

	buf := make([]int, 0)
	token := Token{ make([]string, 0), map[string]int {"Tomek": 0,"Adam": 0, "Krzychu": 0,},buf}
	


	p1 := ProducerConsumer{
		buffor: buffor, 
		bufforMaxsize:5, 
		Monitor:  Monitor{
			id:"Tomek",addr:"127.0.0.1:5555",
			address: adresses,
			RN: map[string]int {"Tomek": 0,"Adam": 0, "Krzychu": 0,},
			haveToken: true,
			token: token ,
		}, 
	}
	p2 := ProducerConsumer{
		buffor: buffor, 
		bufforMaxsize:5, 
		Monitor:  Monitor{
			id:"Adam",addr:"127.0.0.1:5556",
			address: adresses,
			RN: map[string]int {"Tomek": 0,"Adam": 0, "Krzychu": 0,},
			haveToken: false,
			token: token ,
		}, 
	}
	p3 := ProducerConsumer{
		buffor: buffor, 
		bufforMaxsize:5, 
		Monitor:  Monitor{
			id:"Krzychu",addr:"127.0.0.1:5557",
			address: adresses,
			RN: map[string]int {"Tomek": 0,"Adam": 0, "Krzychu": 0,},
			haveToken: false,
			token: token ,
		}, 
	}


	
	

	p1.init()
	p2.init()
	p3.init()
	

	var wg sync.WaitGroup
	wg.Add(603)
	go p1.produce(9,&wg)
	go p1.produce(7,&wg)
	go p1.produce(6,&wg)

	for I := 0; I < 100; I++ {
		go p1.consume(&wg)
		go p3.produce(1,&wg)
		go p2.consume(&wg)
		go p3.produce(2,&wg)
		go p3.consume(&wg)
		go p1.produce(8,&wg)

	}

	wg.Wait()

	log.Println("Synchonizacja!")
	wg.Add(3)
	go p1.read(&wg)
	go p2.read(&wg)
	go p3.read(&wg)
	wg.Wait()

	log.Println(p1.buffor,p2.buffor,p3.buffor)
	if (len(p3.buffor)!=0){
		t.Error("In buffor shoud be 0 elements and are '%i'",len(p2.buffor))
	}
	if (len(p2.buffor)!=0){
		t.Error("In buffor shoud be 0 elements and are '%i'",len(p2.buffor))
	}
	if (len(p1.buffor)!=0){
		t.Error("In buffor shoud be 0 elements and are '%i'",len(p2.buffor))
	}

	p1.DestroyMonitor()
	p2.DestroyMonitor()
	p3.DestroyMonitor()



}