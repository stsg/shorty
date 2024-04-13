# cmd/shorty

В данной директории будет содержаться код, который скомпилируется в бинарное приложение

Sample code to explore content of response object

```
// Explore response object
fmt.Println()
fmt.Println("Response Info:")
fmt.Println("Error:", err)
fmt.Println("Status Code:", resp.StatusCode())
fmt.Println("Status:", resp.Status())
fmt.Println("Proto:", resp.Proto())
fmt.Println("Time:", resp.Time())
fmt.Println("Received At:", resp.ReceivedAt())
fmt.Println("Size:", resp.Size())
fmt.Println("Headers:")
for key, value := range resp.Header() {
    fmt.Println(key, "=", value)
}
fmt.Println("Cookies:")
for i, cookie := range resp.Cookies() {
    fmt.Printf("cookie%d: name:%s value:%s\n", i, cookie.Name, cookie.Value)
}
fmt.Println("Body:\n", resp)
```