package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {

	url := "https://dms-api-stg.tesla.com/api/v1/files/563a5a6e-44c7-42d9-aa87-3be28098cd1d"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6ImYtUDFHLWNjc1hfbjROVkh4aU1lbEhsNElWbyIsImtpZCI6ImYtUDFHLWNjc1hfbjROVkh4aU1lbEhsNElWbyJ9.eyJhdWQiOiJodHRwczovL2Rtcy1zdG9yYWdlLWFwaS1zdGFnZS50ZXNsYS5jb20iLCJpc3MiOiJodHRwOi8vc3NvLWRldi50ZXNsYS5jb20vYWRmcy9zZXJ2aWNlcy90cnVzdCIsImlhdCI6MTY2NDM5NDgwNSwibmJmIjoxNjY0Mzk0ODA1LCJleHAiOjE2NjQzOTg0MDUsImFwcHR5cGUiOiJDb25maWRlbnRpYWwiLCJhcHBpZCI6IjEwYmRjYzcxLTdkYjYtNDViMC1hM2NjLTAxMTVmOGU0MGEzOCIsImF1dGhtZXRob2QiOiJodHRwOi8vc2NoZW1hcy5taWNyb3NvZnQuY29tL3dzLzIwMDgvMDYvaWRlbnRpdHkvYXV0aGVudGljYXRpb25tZXRob2QvcGFzc3dvcmQiLCJhdXRoX3RpbWUiOiIyMDIyLTA5LTI4VDE5OjUzOjI1Ljk4MVoiLCJ2ZXIiOiIxLjAifQ.VFjMX61qLXL5jPgsfwagda4AvHFQWB1qtGTuWLvVt71pXrdBiAQ43eYcNkgMfmsdw2VQjg44mMNbpeBPi5dMuUnanwH-Wgurf2aliZCWw10hnUi4ptphq9p51KaC-_uo-y9tgU4QdHJYpadbEo59tKpCgDIRSgkJ_o4gXpPCLksSZpsMDhIMnwMOU-R-u7Pbx6X-sdO3r-gaBhlBvHqbwNCYxzo2jaizAw7V4B5CLBN-VzugpDUj9zeyhWRcowZ03awCD7ZxiYb5PDXT4HYOuoxejCsHC5QqJ-ZZITEMU1a1c11a7a_j1j4-4fTmLwJ52nShf859QqRzAtv5JI3romSDw1m_C_S4AghRBsPFPQJpN-8MEkcOYPO-1fFz385PR9h74EvBT3Zn_dDUwlgsYCuNjCt0l8pwioeAsa4BCwGR8kejhql-Hsp-aANsB8chvFaGL6lgCn1AR0WAsIqdH4wOq2-Sa9qNWAlfIrco-cEe3KygzVAHuf0xVfOwRSN1C24KzvgBh8xyNHbRaK_GEwak9BFUAg_hP8bRSzGKUqyiqMQqf7wYdxApdOKZ8tg4DXvBDijLsk_73w_EOgTlZ-daJY01ZEwNX_0RkDg2y_ZpYZ2-vGkiGAG4bOq2Fysj_vyfO8vB3J3aB_M_1w4Kc5hdokR8Dt9whUVYMDmfm_4")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	fmt.Println(res.Status)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
