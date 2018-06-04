# Test Helper
Collection of things that can be useful in tests

## Dependencies

This library is dependent on the following revisions of docker components

```
github.com/docker/distribution 8234784a1a66bfee4a6d72d0a3cbd453b7f903d7
github.com/docker/docker 0797e3f61a0402112c00de7b4042fdf3093050a7
github.com/docker/go-connections 988efe982fdecb46f01d53465878ff1f2ff411ce
github.com/docker/go-units 8a7beacffa3009a9ac66bad506b18ffdd110cf97
```

## Run docker in tests
```go
import (
    "fmt"
    "runtime"
	"testing"
	"gitlab.brickchain.com/brickchain/testhelper"
)

func init() {
	err := testhelper.PullImage("nginx")
	if err != nil {
		panic(err)
	}
}

func TestThing(t *testing.T) {
    c, err := testhelper.NewContainer("alpine").WithCmd("/bin/echo test").Start()
    defer c.Remove(true)
    if err != nil {
        t.Fatal(err)
    }
    
    logs, err := c.GetLogs()
    fmt.Println(logs)
}

func TestWithPort(t *testing.T) {
    c, err := testhelper.NewContainer("nginx").Start()
    defer c.Remove(true)
    if err != nil {
        t.Fatal(err)
    }
    
    port, err := c.GetPort("80", "tcp")
    
    fmt.Println("Listening on: localhost:", port)
}

func TestOnMacOrLinux(t *testing.T) {
    c, err := testhelper.NewContainer("nginx").Start()
    defer c.Remove(true)
    if err != nil {
        t.Fatal(err)
    }
    
    var url string
    if runtime.GOOS == "darwin" {
            port, _ := c.GetPort("80", "tcp")
            url = "http://localhost:" + port
    } else {
            ip, _ := c.GetIP()
            url = "http://" + ip + ":80"
    }
    
    fmt.Println("Reach container on:", url)
}

func TestWithLink(t *testing.T) {
    nginx, err := testhelper.NewContainer("nginx").Start()
    defer nginx.Remove(true)
    if err != nil {
        t.Fatal(err)
    }

    c, err := testhelper.NewContainer("alpine").WithLink(nginx.ID, "nginx").WithCmd("/bin/ping -c 1 nginx").Start()
    defer c.Remove(true)
    if err != nil {
        t.Fatal(err)
    }

    code, err := c.Wait()
    if err != nil {
        t.Fatal(err)
    }

    if code != 0 {
        logs, err := c.GetLogs()
        if err != nil {
            t.Fatal(err)
        }

        t.Fatal(logs)
    }
}

func TestWithEnv(t *testing.T) {
	env := []string{
		"TEST=stuff",
	}
	c, err := testhelper.NewContainer("alpine").WithEnv(env).WithCmd("/usr/bin/env").Start()
	defer c.Remove(true)
	if err != nil {
		t.Fatal(err)
	}
	
	c.Wait()
	
    logs, err := c.GetLogs()
    if err != nil {
        t.Fatal(err)
    }
    
    fmt.Println(logs)
}
```

## Helper for TestMain
```go
package things
import (
	"gitlab.brickchain.com/brickchain/testhelper"
	"time"
	"testing"
)

var containers map[string]*testhelper.Container = make(map[string]*testhelper.Container)

func TestMain(m *testing.M) {
	containers["ipfs"] = testhelper.NewContainer("registry.brickchain.com/3rd-party/go-ipfs:master").
					WithEnv([]string{"CLEAR_BOOTSTRAP=yes"})
	defer containers["ipfs"].Remove(true)

	containers["redis"] = testhelper.NewContainer("redis")
	defer containers["redis"].Remove(true)

	err := testhelper.SetupContainers(containers)
	if err != nil {
		panic(err)
	}

	// wait while everything starts
	time.Sleep(time.Second*2)

	m.Run()
}

func getAddr(name, port string) string {
	endpoint, _ := testhelper.GetEndpoint(containers[name], port)
	return endpoint
}

func TestThing(t *testing.T) {
        DoThings(getAddr("redis", "6379"))
}
```
