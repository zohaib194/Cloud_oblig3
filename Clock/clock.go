package clock


func main() {
	ticker := time.NewTicker(time.Second * 120)
	go func() {
		for t := range ticker.C {
			//call functions
			fmt.Printf("\n", t)
			InvokeWebhook()
			//GetFixerSevenDays(time.Now().AddDate(0, 0, -7), time.Now())
		}
	}()
	ticker.Stop()
}