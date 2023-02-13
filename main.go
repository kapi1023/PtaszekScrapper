package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

type ProductSkuData struct {
	Results []struct {
		DefaultVariant struct {
			Sku string `json:"sku"`
		} `json:"defaultVariant"`
	} `json:"results"`
}

type DataProduct struct {
	CategoryName   string `json:"categoryName"`
	DefaultVariant struct {
		ID          int    `json:"id"`
		BrandName   string `json:"name"`
		Unit        string `json:"unit"`
		PackageInfo struct {
			PackageSize float64 `json:"packageSize"`
		} `json:"packageInfo"`
	} `json:"defaultVariant"`
}

func main() {
	authCode := "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJhdWQiOiIzMWMyZWQ3N2QyYTQ3YWM4YTAxZDc4MDMwZWFhMWEyYyIsImp0aSI6ImM5MDNmZGMwZjUzNDRiM2U1MjZlMWFlY2JjM2U5YTdkNmFmMmJiZDgxNWE2YTI1ZTBjZjljYWI3NmYzNTE2NjI4YjBmMThjMzgxMWY5ZTg2IiwiaWF0IjoxNjc2Mjg1NDY4LjM2MTgyOSwibmJmIjoxNjc2Mjg1NDY4LjM2MTgzNywiZXhwIjoxNjc2MzcxODY4LjMzODY3Miwic3ViIjoiYW5vbl9kNzQyMmE5Zi0yNDYwLTQ0MzEtOWVhYy1jODA5YTMzOGNkNzUiLCJzY29wZXMiOltdfQ.GqIFbLddGPQh_oBw5JwQ7riVkpBsW64eq5s6QxfC7029mdzjIao3JHlkGRpS3a3UAiWhcp3kopNhRpFi6suLI9XSxNBrpYibckYWDf-hDODdlNReoW7CZJs8EuXPtMxVWdw9HpjqKJZ-XHFf47Uocb1-9UWQR_JDVGQ0A1ayqxE_XN9HvrDdJmz4YYIcF_U0XRjFYY2P-l_6y-5fybN5WqUZeFx8SdXXH6ZuVC7QQeOEqUA_8xiuMFgrPFf_u_XIZ7cArnymk3UZ_7Ch2QWMHkHiq2nKGViq4eMreMbsI4RJIIPa4DPIupU0A3f5zXmVhGE4zgtPbyupwYzVOweEEA"
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
		var dataProductStruct DataProduct
		file, err := os.OpenFile("products_data.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()

		writer := csv.NewWriter(file)

		err = json.Unmarshal([]byte(body), &dataProductStruct)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return
		}
		fmt.Println("BrandName: ", dataProductStruct.DefaultVariant.BrandName, "PackageSize: ", dataProductStruct.DefaultVariant.PackageInfo.PackageSize, "Unit: ", dataProductStruct.DefaultVariant.Unit)

		brandName := dataProductStruct.DefaultVariant.BrandName
		categoryName := dataProductStruct.CategoryName
		packageSize := dataProductStruct.DefaultVariant.PackageInfo.PackageSize
		unit := dataProductStruct.DefaultVariant.Unit

		err = writer.Write([]string{
			brandName, categoryName, strconv.FormatFloat(packageSize, 'f', -1, 64), unit,
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
