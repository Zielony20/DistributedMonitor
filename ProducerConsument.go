package main

import( 
	"log"
	//"time"	
	"sync"
)


type ProducerConsumer struct {
	
	buffor []int
	bufforMaxsize int // max size
	Monitor	
}

func (p *ProducerConsumer) init() {
	p.initMonitor()
	
	p.buffor = []int{}
	p.bufforMaxsize = 5

}


func (p *ProducerConsumer) read(wg *sync.WaitGroup) {
	
	p.lock()
	
	for !p.haveToken{
		p.wait()
	}
	p.buffor = p.token.buffor	
	log.Printf("'%s': Synchronized buffor value",p.id)
	
	p.unlock()
	wg.Done()
}


func (p *ProducerConsumer) produce(i int, wg *sync.WaitGroup){
	p.lock()

	for !p.haveToken || len(p.token.buffor) >= p.bufforMaxsize{
		p.wait()
	}
	
	p.token.buffor = append(p.token.buffor,i)
	log.Println(p.token.buffor)
	p.buffor = p.token.buffor
	log.Printf("'%s' Buffor update - \033[32m Produce \033[0m",p.id)
	p.notifyAll()
	
	p.unlock()
	wg.Done()

}

func (p *ProducerConsumer) consume(wg *sync.WaitGroup){
	
	p.lock()
	
	
	for !p.haveToken || len(p.token.buffor)==0{
		p.wait()
	}
	
	log.Println(p.token.buffor)
	p.token.buffor = p.token.buffor[1:]
	log.Println(p.token.buffor)
	p.buffor = p.token.buffor
	log.Printf("'%s' Buffor update - \033[32m Consume \033[0m",p.id)
	p.notifyAll()

	p.unlock()

	wg.Done()
}

