# plotter

Pipe your data from shell to graph.

## try it out
* `$ gb build`
* `$ ./bin/plotter`
* Open browser tab pointing at http://localhost:7272
* try example below

## line charts
```
$ { cat <<EOF
0,2,8
1,10,-3
2,11,22
3,67,9
4,90,23
5,20,-87
6,20,45
7,23,2
8,40,9
9,32,22
EOF
} | plot
```

## histograms
```
$ echo '1,1,1,3,6,4,4,8,2,2,2,8,9' | tr , '\n' | plot
```
