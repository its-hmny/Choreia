# Choreia

A static analyzer to generate Choreography Automata from Go source.

## What is

Aim of this project is to write a complete static analyzer that will parse and extract metadata from a Go source file and from that extrapolate a Choreography Automata, that is a Finite State Automta (FSA) that represents how the different Goroutines interact with each other during the program execution flow.

Some of the use cases of such program could be:

- _Debugging pouposes_: The program will output a SVG file that represents Choreography Automata and this can be a visual aid to debugging complex distributed systems
- _Correctness and well-formedness_: If the final Choreography Automata fullfills some properties such as Correctness and Well-formedness that in turn allows for some assumption to be made about our program and the interactions occuring between its parts
- _Visual interpretation of protocols_: This tool can be used to visualize both standard and proprietary comunication-based protocol, to do so we need only a mock implementation of our protocol in Go and the tool will take care of the visualization

## How to

Before running the project you need to install some dependencies with the `go get` command, after that you can run the project with:

```console
usr@computer:~/Choreia$ go run choreia/main.go -i input_file.go
```

or optionally build it and then running it as a standalone executable:

```console
usr@computer:~/Choreia$ go build o your_path choreia/main.go
usr@computer:~/Choreia$ ./your_path choreia/main.go -i input_file.go
```

the latter is especially indicated when parsing large files as the compilation increases the execution speed.
In addition `-i` arg to indicate the input file, other CLI argument are the following:

| Shorthand | Extended   | Usage                                                 | Default         |
| :-------- | :--------- | :---------------------------------------------------- | :-------------- |
| `-i`      | `--input`  | The path to the Go source file                        |
| `-o`      | `--output` | The path to where the data wll be saved               | `./choreia_out` |
| `-t`      | `--trace`  | Prints to the stdout a trace of the AST while parsing |
| `-h`      | `--help`   | Show help message and usage instructions              |

## Credits & Licensing

This project was made by [me](https://github.com/its-hmny) as Bachelor's degree Thesis for the Computer Science course at University of Bologna.
Special thanks to the professor [Ivan Lanese](https://github.com/lanese) and to each contributors of the library used as dependencies.
