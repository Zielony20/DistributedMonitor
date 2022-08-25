package main

import( 
	"strings"
	"strconv"	
	//"log"
)


type Token struct {
    queue []string `default:make([]string, 0)`
	LN map[string]int `default:{"1": 0,"2": 0, "3": 0,}`
	buffor []int
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

	result += "B:"
	if len(q.buffor) != 0 {
		for _, i := range q.buffor {
			result = result + strconv.Itoa(i) + "$"
		}
	}
	

	return result

}


func (q *Token) stringToToken(str string){
//"1--0$2--5$3--0$Q:1$2$B:1$3$6$1$"

	q.LN = make(map[string]int,0)
	q.queue = make([]string,0)
	q.buffor = make([]int,0)
	endOfLN := strings.LastIndex(str, "Q:")
	endOfQueue := strings.LastIndex(str, "B:")
	first_part := strings.Split(str[0:endOfLN-1], "$")


	for _,v := range first_part{
		id_and_lvalue := strings.Split(v, "--")
		q.LN[id_and_lvalue[0]], _ = strconv.Atoi(id_and_lvalue[1])
	}

	
	second_part := strings.Split(str[endOfLN+2:endOfQueue-1], "$")

	for _, i := range second_part{
		q.EnqueueToken(i)
	}

	// if queue is not empty
	if( endOfQueue+2 != len(str)){
		third_part := strings.Split(str[endOfQueue+2:len(str)-1], "$")
	
		for _,i := range third_part{
			
			temp, _ := strconv.Atoi(i)
			q.buffor = append(q.buffor,temp) 
		}
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

