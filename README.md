dxf-go
======

A go package for reading and writing DXF CAD files.

[![Build Status](https://dev.azure.com/ixmilia/public/_apis/build/status/dxf-go?branchName=master)](https://dev.azure.com/ixmilia/public/_build/latest?definitionId=28)

## Usage

Acquisition:

``` bash
go get -d -t github.com/ixmilia/dxf-go
go generate github.com/ixmilia/dxf-go
```

Open a DXF file:

``` go
import dxf "github.com/ixmilia/dxf-go"
// ...

drawing, err := dxf.ReadFile("path/to/file.dxf")
// if err != nil

for _, entity := range drawing.Entities {
    switch ent := entity.(type) {
    case *dxf.Line:
        fmt.Printf("Line from %s to %s\n", ent.P1.String(), ent.P2.String())
    // ...
    }
}
```

Save a DXF file:

``` go
import dxf "github.com/ixmilia/dxf-go"
//...

drawing := *dxf.NewDrawing()
line := dxf.NewLine()
line.P1 = dxf.Point{X: 1.0, Y: 2.0, Z: 3.0}
line.P2 = dxf.Point{X: 4.0, Y: 5.0, Z: 6.0}
drawing.Entities = append(drawing.Entities, line)
// ...

err := drawing.SaveFile("path/to/file.dxf")
// if err != nil
```
