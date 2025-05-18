# control #

Is a simple tool for teaching the basics of classic control theory.
It is intended to be used by students to create Nyquist plots, 
Bode plots and root locus analysis plots (Evans plots). It is possible 
to simulate a linear system and to create a step response plot. 
It is also possible to calculate the response of a system to an arbitrary 
input signal.
It works on Android and iOS tablets, Windows, Linux and MacOS.
Although it is possible to run it on a smartphone, it is not a good 
platform because the screen is too small.
There is a [static web page](https://hneemann.github.io/control/) 
that can be used. But that static page does not allow storing scripts on 
the server. 
If the application is operated with a backend, there are the functions 
`Save`, `Save as` and `Open` to store scripts.

# Usage #

The help icon offers some examples and shows how to use it. Error messages
offer a description of the error and a hint how to fix it.

# Implementation #

The engine is based on this [parser](https://github.com/hneemann/parser2).
There are some additional first-class types that can be used. One is the 
polynomial. Polynomials can be added, subtracted and multiplied. If two 
polynomials are divided, the result is a linear system as a further 
first-class type. Linear systems can also be added, subtracted, 
multiplied and divided. There is the constant `s` which corresponds to a 
polynomial `a*x+b`, where `b=0` and `a=1`. This way it is possible to
create a linear system by simply typing `(s+2)/((s+3)*(s+1))` which 
feels very natural.

The parser itself generates a single result value, which is output as HTML. 
The result can be a list or a set. Graphical images can also be a result 
value and are displayed as SVG. They are also additional first-class types 
that can be contained in lists, for example, if several graphics are to be 
output.

The parser is compiled to WEB-Assembly and runs in the browser. 
The script, which serves as input for the parser, does not leave 
the browser. The APP can therefore be operated as a static WEB page.

# Limits #

The simulation of linear systems is for now done by using the Euler method, with a 
fixed number of steps which is good enough for most simple cases, but not 
sufficient for simulating more complex systems.

