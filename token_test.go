package main

import "testing"


var queue = make([]string, 0)
var LN = map[string]int{"1": 0,}
var buffor = make([]int, 0)

var token = &Token{queue,LN,buffor}



func TestTokenEmpty1(t *testing.T) {

    token.EnqueueToken("2")
    if token.EmptyToken(){
        t.Error("Token empty")    
    }

}

func TestConvertionStringToTokenAndBack(t *testing.T) {
    
    str := "1--0$2--5$3--0$Q:1$2$B:1$3$6$1$"
    
    buf := [4]int {1,3,6,1}
    //var name_of_array [5] int {23 ,12 ,44}
    token.stringToToken(str)
    
            
    for j,i := range token.buffor{
        if i != buf[j]{
            t.Error("Wrong value in buffor",i)
        }
    }
    str = "1--0$2--5$3--0$Q:1$2$B:1$3$6$3$"
    token.buffor = []int {1,3,6,3}
    //temp := token.TokenToString()
    //if (temp != str){
    //    t.Error("Wrong value in string:",temp,"\n shoud be:",str)
    //}

}









