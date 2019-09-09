package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/tealeg/xlsx"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	excelFileName := "XXX.xlsx"// 存放房源网站网站列表的Excel表格
	xlFile, err := xlsx.OpenFile(excelFileName)

	if err != nil {
	fmt.Println("open file error : ", err)
		return
	}

	target := xlsx.NewFile()//目标Excel文件
	ts2, err := target.AddSheet("Sheet2")
	if err != nil {
		fmt.Println("new sheet2 error: ", err)
		return
	}
	sheet2 := xlFile.Sheet["Sheet2"]
	first := ts2.AddRow()
	first.AddCell().SetString("PC")
	first.AddCell().SetString("Level1")
	first.AddCell().SetString("Level2")
	first.AddCell().SetString("Level3")

	lss := make(LocatiionStructs, 0)

	wg := sync.WaitGroup{}

	for index, row := range sheet2.Rows {
		if index % 100 == 0 {
			time.Sleep(5 * time.Second)//防止超时
		}
		link := row.Cells[0].String()
		if strings.Contains(link, "http"){
			wg.Add(1)
			go func(index int, link string) {
				lss = append(lss,scrape(index, link))
				wg.Done()
				time.Sleep(100)// 不能同时启动太多线程，会超时
			}(index, link)
		}

	}

	wg.Wait()
	sort.Sort(lss)

	fmt.Println("size is : ", lss.Len())


	for _, ls := range lss {
		if ls.URL != ""{
			ts2r := ts2.AddRow()
			ts2r.AddCell().SetString(ls.URL)
			ts2r.AddCell().SetString(ls.Location1)
			ts2r.AddCell().SetString(ls.Location2)
			ts2r.AddCell().SetString(ls.Location3)
		}
	}

	if err := target.Save("target.xlsx"); err != nil {
		fmt.Println("save target error : ", err)
		return
	}
	fmt.Println("success save file ")
}

//使用原顺序，所以需要排序
type LocationStruct struct {
	 index int
	 URL string
	 Location1 string
	 Location2 string
	 Location3 string
}

type LocatiionStructs []LocationStruct

func (lss LocatiionStructs)Len() int{
	return len(lss)
}
func (lss LocatiionStructs)Less(i, j int) bool{
	return lss[i].index < lss[j].index
}
func (lss LocatiionStructs)Swap(i, j int){
	lss[i], lss[j] = lss[j], lss[i]
}

func scrape(index int, url string) LocationStruct{
	res, err := http.Get(url)

	if err != nil {
		fmt.Println(url, "---",err)
		return LocationStruct{}
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		fmt.Println("status code is not 200: ",url)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("parse doc error: ", err)
		return LocationStruct{}
	}

	ls := LocationStruct{index:index, URL:url}

	// selector
	doc.Find(".breadcrumbs a").Each(func(i int, selection *goquery.Selection) {

		if i == 0 {
			fmt.Printf("url: %s, index : %d, %s \n", url, i, selection.Text())
			ls.Location1 = selection.Text()
		}
		if i == 1 {
			fmt.Printf("url: %s, index : %d, %s \n", url, i, selection.Text())
			ls.Location2 = selection.Text()
		}
		if i == 2 {
			fmt.Printf("url: %s, index : %d, %s \n", url, i, selection.Text())
			ls.Location3 = selection.Text()
		}
	})

	return ls
}
