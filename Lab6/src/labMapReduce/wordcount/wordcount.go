package main

import (
	"labMapReduce/mapreduce"
	"hash/fnv"
	"unicode" 	
	"strconv"
	"strings"
)

func mapFunc(input []byte) (result []mapreduce.KeyValue) {
	var text = string(input)
	text = text + " " //bizu pra entrar a última palavra
	text = strings.ToLower(text)
	var word string = ""
	var mapAux map[string]int = make(map[string]int)

	for _, c := range text {  
		if !unicode.IsLetter(c) && !unicode.IsNumber(c){
			mapAux[word]++
			word = ""
		} else {
			word = word + string(c)
		}
	}
	
	for k, v := range mapAux {
		result = append(result, mapreduce.KeyValue{k,strconv.Itoa(v)})
	}
	return result	
}

func reduceFunc(input []mapreduce.KeyValue) (result []mapreduce.KeyValue) {
	myMap := make(map[string]int)
			
	for _, kv := range input {
		k := kv.Key
		v,_ := strconv.Atoi(kv.Value)
		myMap[k] += v
	}

	for k, v := range myMap {
		result = append(result, mapreduce.KeyValue{k,strconv.Itoa(v)})
	}

	return result	
}

func shuffleFunc(task *mapreduce.Task, key string) (reduceJob int) {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() % uint32(task.NumReduceJobs))
}
