dxf-go
======

A go package for reading and writing DXF CAD files.

## Usage

Acquisition:

``` bash
go get github.com/IxMilia/dxf-go
go generate github.com/IxMilia/dxf-go
```

Open a DXF file:

``` go
import dxf "github.com/IxMilia/dxf-go"
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
import dxf "github.com/IxMilia/dxf-go"
//...

drawing := *dxf.NewDrawing()
line := dxf.NewLine()
line.P1 = dxf.Point{1.0, 2.0, 3.0}
line.P2 = dxf.Point{4.0, 5.0, 6.0}
drawing.Entities = append(drawing.Entities, line)
// ...

err := drawing.SaveFile("path/to/file.dxf")
// if err != nil
```
