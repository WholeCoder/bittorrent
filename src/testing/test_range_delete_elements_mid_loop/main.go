package main

import "fmt"

func RemoveAtIndex(s []int, index int) ([]int,int) {
    ret_value := s[index]
    ret := make([]int, 0)
    ret = append(ret, s[:index]...)
    return append(ret, s[index+1:]...),ret_value
}

func main() {
    lst2 := []int{}
    lst := []int{0, 1, 2, 3, 4}
    for index,piece := range lst {
        lst, piece = RemoveAtIndex(lst, index)
        fmt.Printf("%#v\n", lst)
        lst2 = append(lst2, piece)
    }
}
