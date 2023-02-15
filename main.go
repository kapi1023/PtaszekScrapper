package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type ProductSkuData struct {
	Results []struct {
		DefaultVariant struct {
			Sku string `json:"sku"`
		} `json:"defaultVariant"`
	} `json:"results"`
}

type ProductData struct {
	Id             int    `json:"id"`
	CategoryId     string `json:"CategoryId"`
	BrandName      string `json:"brandName"`
	CategoryName   string `json:"categoryName"`
	DefaultVariant struct {
		Name        string `json:"name"`
		Sku         string `json:"sku"`
		ProductId   int    `json:"productId"`
		Unit        string `json:"unit"`
		PackageInfo struct {
			PackageUnit string  `json:"packageUnit"`
			PackageSize float64 `json:"packageSize"`
		} `json:"packageInfo"`
		ItemVolumeInfo string `json:"itemVolumeInfo"`
		Media          struct {
			Images    []string `json:"images"`
			MainImage string   `json:"mainImage"`
			ListImage string   `json:"listImage"`
		}
	} `json:"defaultVariant"`
}

func main() {
	authCode := "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJhdWQiOiIzMWMyZWQ3N2QyYTQ3YWM4YTAxZDc4MDMwZWFhMWEyYyIsImp0aSI6ImY4NDg1NTM2ZTcyNWYwYWJhYTA2YTViNWIzYzY2MjhjZWY1ODE2Y2I1ZDMyZDhhMzQxY2M3YzdlNGM1ZmFlN2M4ZDM5NGQ4YjhhZTMzMWVmIiwiaWF0IjoxNjc2NDkyOTE5LjEzNzg5OCwibmJmIjoxNjc2NDkyOTE5LjEzNzkwNSwiZXhwIjoxNjc2NTc5MzE5LjExODMzNiwic3ViIjoiYW5vbl82YTI1ZjUyYi01MWIyLTRkZGQtODFjNC00ZDM2ODI3NzJiNWUiLCJzY29wZXMiOltdfQ.qKRGlJfuJXLUsciguhsJeuD-GqXMVAQsXm5PpW5wntBqpyqlwvsT2efdic0N7v0gymLrz6GAPK7rzaqzTTDFoZJwxVKP3hnBIIDJV0gLC0JEVabHDtUy_JwQSs7LUAZtLb_wUlvowkJpVR3ZKyHBKBF0rEXWRQ9T2IP8wXuYs0GtZVLGcBJmeqTtLoM2cLrI5Ddkx_AfLAhYVJJxG72TKIqGQbGU-LQ3vpoaDM6ewsEJIl4O8jvCCJpQJJe9Y8Iv05_EaBYJhtL4jh-OQRfTXgqG65IoJ0IcOq1Bq7TQd_9SrTSonQeRgVSyxuYqX23Y8UOWedG9TpFyfQIBwlrvMg"
	file, err := os.Open("categorieId.csv")
	if err != nil {
		fmt.Println("Nie udało się otworzyć pliku")
		return
	}
	defer file.Close()
	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Nie udało się wczytać danych")
		return
	}
	for _, record := range records {
		fmt.Println(record[0])
		categoryId := record[0]
		getProductId(authCode, categoryId)
		getProductData(authCode)
	}

}

func getProductData(authCode string) {

	file, err := os.Open("productId.csv")
	if err != nil {
		fmt.Println("Nie udało się otworzyć pliku")
		return
	}
	defer file.Close()
	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Nie udało się wczytać danych")
		return
	}
	for _, record := range records {
		fmt.Println(record[0])
		sku := record[0]

		url := "https://zakupy.auchan.pl/api/v2/products/sku/" + sku + "?hp=pl"
		fmt.Println(url)
		client := &http.Client{}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
		req.Header.Set("Authorization", authCode)

		res, err := client.Do(req)
		if err != nil {
			println(err.Error())
			os.Exit(2)
		}

		println("Status code:", res.StatusCode)
		body, error := ioutil.ReadAll(res.Body)
		if error != nil {
			fmt.Println(error)
		}
		res.Body.Close()
		var productDataStruct ProductData
		file, err := os.OpenFile("products_data.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()

		writer := csv.NewWriter(file)

		err = json.Unmarshal([]byte(body), &productDataStruct)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return
		}
		images := ""
		fmt.Println("BrandName: ", productDataStruct.BrandName, "PackageSize: ", productDataStruct.DefaultVariant.PackageInfo.PackageSize, "Unit: ", productDataStruct.DefaultVariant.Unit)
		for _, image := range productDataStruct.DefaultVariant.Media.Images {

			images += image + ","
		}
		images = strings.TrimSuffix(images, ",")
		err = writer.Write([]string{
			strconv.Itoa(productDataStruct.Id),
			productDataStruct.CategoryId,
			productDataStruct.BrandName,
			productDataStruct.CategoryName,
			productDataStruct.DefaultVariant.Name,
			productDataStruct.DefaultVariant.Sku,
			strconv.Itoa(productDataStruct.DefaultVariant.ProductId),
			productDataStruct.DefaultVariant.Unit,
			productDataStruct.DefaultVariant.PackageInfo.PackageUnit,
			strconv.FormatFloat(productDataStruct.DefaultVariant.PackageInfo.PackageSize, 'f', -1, 64),
			productDataStruct.DefaultVariant.ItemVolumeInfo,
			images,
			productDataStruct.DefaultVariant.Media.MainImage,
			productDataStruct.DefaultVariant.Media.ListImage,
		})
		if err != nil {
			panic(err)
		}

		writer.Flush()
	}
}

func getProductId(authCode string, categoryId string) {
	url := "https://zakupy.auchan.pl/api/v2/products?categoryId=" + categoryId + "&itemsPerPage=999&page=1&hl=pl"
	client := &http.Client{}
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	req.Header.Set("Authorization", authCode)

	res, err := client.Do(req)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	println("Status code:", res.StatusCode)
	body, error := ioutil.ReadAll(res.Body)
	if error != nil {
		fmt.Println(error)
	}
	res.Body.Close()

	var dataStruct ProductSkuData

	file, err := os.OpenFile("productId.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	err = json.Unmarshal([]byte(body), &dataStruct)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	for _, result := range dataStruct.Results {
		i := result.DefaultVariant.Sku
		err := writer.Write([]string{
			i,
		})
		fmt.Println(i)
		if err != nil {
			panic(err)
		}
	}
	writer.Flush()
}
