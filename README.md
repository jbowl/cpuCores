playing around with concurrency

./cpuCores -Fn=40 -workers=2  displays first 40 fib numbers using 2 goroutines

...
...
...
165580141        CPU USAGE: 31.31 100.00 98.99 100.00 
267914296        CPU USAGE: 31.31 100.00 98.99 100.00 
433494437        CPU USAGE: 20.20 100.00 100.00 100.00 
701408733        CPU USAGE: 17.17 100.00 100.00 100.00 
1134903170       CPU USAGE: 19.00 100.00 100.00 100.00 
1836311903       CPU USAGE: 18.00 100.00 100.00 100.00 
2971215073       CPU USAGE: 9.00 100.00 100.00 10.10 
4807526976       CPU USAGE: 8.00 6.93 100.00 5.00 
calculated 49 Fn numbers using 3 goroutines in 37.701888812s
 
