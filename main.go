package main

func main()  {
	a := &App{}
	a.Initialize(GetEnv())
	//fmt.Println(a)
	a.Run(":8000")
}
