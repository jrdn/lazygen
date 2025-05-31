# LazyGen
LazyGen is a template-based code generator using go templates in comments to generate repetitive code
rather than copy/paste/search/replace. 

## Using lazygen
### Templates
First you need to specify a template. A template takes the format:
```
/* lazygen:template NAME
BODY
*/
````
Each template is given a name which is used later to specify which template you want to instantiate.

### Instances
Instances fill out the template with arguments provided.
```
// lazygen:instance NAME k=v k2=v2
``` 

### Functions
Sprig functions are available for use in templates.

https://masterminds.github.io/sprig/

For example:
```
/* lazygen:template HelloWorld
func {{title .h}}{{title .w}}() {
    fmt.Println("{{upper .h}} {{.w}}")
}
/*

// lazygen:instance h=hello w=world
```
would render to:
```
func HelloWorld() {}
    fmt.Println("Hello world")
}
```

### Running the generator
At the top of the file you'd like to generate, add the go generate comment:
```
//go:generate lazygen $GOFILE
```
then, in the directory containing the file, run `go generate`