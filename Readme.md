### GPU-Z shared memory reader

This repo is used for getting data and sensors values from running GPU-Z utility by shared memory. 

**Running GPU-Z is required** while using this code.

### Example of usage

```go
package main

import (
	"fmt"
	"log"
	
	"github.com/hotafrika/gpuz-reader"
)

func main() {
	sm := gpuz.DefaultSharedMemory()
	stat, err := sm.GetStat()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(stat.GetAvailableRecords())
	fmt.Println(stat.GetAvailableSensors())
	fmt.Println(stat.GetRecord("CardName"))
	fmt.Println(stat.GetSensor("GPU Load"))
	fmt.Println(stat.GetSensorValue("GPU Temperature"))
}
```
