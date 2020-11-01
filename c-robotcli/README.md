# Robot Simulator Command Line Interface

We wish to create a simple Golang application which allows the use of the `librobot` simulator from a command line for testing purposes.

Use whatever Golang libraries you see fit.

## Requirements

### Part One

Create an interactive REPL type command line application which simulates a warehouse containing one or more robots, and allows issuing of tasks to these robots interactively.

The application should provide a prompt where the user is able to enter a task for a robot in string form.  The state of the simulated environment should be persisted between tasks being entered.

Provide relevant user documentation to allow the use of the application.

### Part Two

Add some kind of print out representation of the state of the simulation to the CLI application, which allows the user to see the simulation evolving in real time.
