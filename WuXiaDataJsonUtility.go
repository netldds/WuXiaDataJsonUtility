package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type RideHorse_Accumulation struct {
	RideHorse_Accumulation    string
	Acceleration_Accumulation string
}

type PlayerID struct {
	PlayerID                string
	Damage_Accumulation     string
	Kill_Accumulation       string
	Multi_Kill_Accumulation string
	Trip                    string
}

type WeaponID struct {
	WeaponID                   string
	Assassination_Accumulation string
}

type BufferID struct {
	BufferID     string
	Accumulation string
}
type ItemID struct {
	ItemID                string
	Loot_Accumulation     string
	Activity_Accumulation string
	PlayerID
}

type WuXiaJson struct {
	Table_Horses_Ride_Acticity       RideHorse_Accumulation
	Table_Player_Data                []PlayerID
	Table_Dagger_Assassination       WeaponID
	Table_Killing_ID                 []BufferID
	Table_Skills_Items_Loot_Activity []ItemID
}

const (
	PATH = "E:/WuXiaData/All"

	PATHCSV = "E:/GOProject/src/WuXiaDataJsonUtility/Bin/所有ID的拾取和使用数据.csv"

	PATHCSVDEATH = "E:/GOProject/src/WuXiaDataJsonUtility/Bin/死亡原因.csv"

	PATHCSVPLAYER = "E:/GOProject/src/WuXiaDataJsonUtility/Bin/账号数据.csv"

	//ITEM		物品ID-捡取累计-使用累计
	//DR			死亡ID-次数累计
	//PLAYER	用户数据
	HandleType = "PLAYER"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func GetFileslist(f func(fileName string) bool) []os.FileInfo {
	list, err := ioutil.ReadDir(PATH)
	checkErr(err)
	var retList []os.FileInfo
	for _, v := range list {
		ext := filepath.Ext(v.Name())
		ok := f(ext)
		if ok {
			retList = append(retList, v)
		}
	}
	return retList
}

func Processing(fs os.FileInfo, wg *sync.WaitGroup, ch chan []itemData) {

	absPath := PATH + "/" + fs.Name()
	resourceByte, err := ioutil.ReadFile(absPath)
	checkErr(err)
	resJson := WuXiaJson{}
	json.Unmarshal(resourceByte, &resJson)
	//fmt.Println(resJson)
	reshandle(resJson, ch)
	defer func() {
		wg.Done()
		fmt.Printf("%-50s was done\n", fs.Name())
	}()
}

type itemData struct {
	ID, Alpha, Beta, Gamma int
}

func elemntInSlice(data itemData, itemData []itemData) (bool, int) {
	for i, v := range itemData {
		if v.ID == data.ID {
			return true, i
		}
	}
	return false, -1
}
func reshandle(resJson WuXiaJson, ch chan []itemData) {
	ret := [][]string{
		{"ID", "拾取次数", "使用次数"},
	}

	idata := make([]itemData, 1)

	switch HandleType {
	case "DR":
		for _, v := range resJson.Table_Killing_ID {
			index, err := strconv.Atoi(v.BufferID)
			accumulation, err2 := strconv.Atoi(v.Accumulation)
			if err != nil || err2 != nil {
				panic("Json Read error")
			}
			if len(idata) == 0 {
				idata = append(idata, itemData{ID: index, Alpha: accumulation, Beta: 0})
				ret = append(ret, []string{
					strconv.Itoa(index),
					strconv.Itoa(accumulation),
					strconv.Itoa(0)})
			}
			temporary := itemData{ID: index, Alpha: accumulation, Beta: 0}
			ok, n := elemntInSlice(temporary, idata)
			if !ok {
				idata = append(idata, temporary)
				ret = append(ret, []string{
					strconv.Itoa(index),
					strconv.Itoa(accumulation),
					strconv.Itoa(0)})
			} else {
				idata[n].Alpha += temporary.Alpha
				idata[n].Beta += temporary.Beta
			}
		}
	case "ITEM":
		for _, v := range resJson.Table_Skills_Items_Loot_Activity {
			index, err := strconv.Atoi(v.ItemID)
			lootN, err2 := strconv.Atoi(v.Loot_Accumulation)
			ActivityN, err3 := strconv.Atoi(v.Activity_Accumulation)
			if err != nil || err2 != nil || err3 != nil {
				panic("Json read error.")
			}

			//第一次
			if len(idata) == 0 {
				idata = append(idata, itemData{ID: index,
					Alpha: lootN,
					Beta: ActivityN})
				ret = append(ret, []string{strconv.Itoa(index),
					strconv.Itoa(lootN),
					strconv.Itoa(ActivityN)})
			}

			temporary := itemData{ID: index, Alpha: lootN, Beta: ActivityN}

			//不存在，则加入
			ok, n := elemntInSlice(temporary, idata)
			if !ok {
				idata = append(idata, temporary)
				ret = append(ret, []string{strconv.Itoa(index),
					strconv.Itoa(lootN),
					strconv.Itoa(ActivityN)})
			} else {
				idata[n].Alpha += temporary.Alpha
				idata[n].Beta += temporary.Beta
			}

		}
	case "PLAYER":
		for _, v := range resJson.Table_Player_Data {
			id, err := strconv.Atoi(v.PlayerID)
			alpha, err2 := strconv.Atoi(v.Damage_Accumulation)
			beta, err3 := strconv.Atoi(v.Kill_Accumulation)
			gamma, err4 := strconv.Atoi(v.Trip)
			if err != nil || err2 != nil || err3 != nil || err4 != nil {
				panic("Json read error.")
			}
			idata = append(idata, itemData{ID: id, Alpha: alpha, Beta: beta, Gamma: gamma})
		}
	default:
		fmt.Println("None of Handle Type was selected (ITEM,DR),Correct  Name needed!")
	}

	ch <- idata
}

func main() {
	wg := sync.WaitGroup{}

	fileList := GetFileslist(func(fileName string) bool {
		if fileName == ".json" {
			return true
		}
		return false
	})
	totalChannel := make(chan []itemData, len(fileList))

	fmt.Println(cap(totalChannel))
	fmt.Printf("A TOTAL OF %d FIELS\n", len(fileList))
	wg.Add(len(fileList))
	//wg.Add(1)
	for _, v := range fileList {
		go Processing(v, &wg, totalChannel)
	}
	//go Processing(fileList[0], &wg)
	wg.Wait()

	close(totalChannel)

	ret := [][]string{
		{"ID", "Alpha", "Beta","Gamma"},
	}
	totalData := make([]itemData, 1)

	for v := range totalChannel {
		for _, vi := range v {
			//第一次
			temporary := itemData{ID: vi.ID, Alpha: vi.Alpha, Beta: vi.Beta,Gamma:vi.Gamma}
			if len(totalData) == 0 {
				totalData = append(totalData, temporary)
			}
			//使用玩家数据时，永远是假，ID有一样的，但数据不能累加
			ok, n := elemntInSlice(temporary, totalData)
			ok=false
			if !ok {
				totalData = append(totalData, temporary)
			} else {
				totalData[n].Alpha += temporary.Alpha
				totalData[n].Beta += temporary.Beta
				totalData[n].Gamma += temporary.Gamma
			}

		}
	}
	//写入到[][]string
	for _, vi := range totalData {
		ret = append(ret, []string{
			strconv.Itoa(vi.ID),
			strconv.Itoa(vi.Alpha),
			strconv.Itoa(vi.Beta),
			strconv.Itoa(vi.Gamma),
		})

	}
	var fpath string
	switch HandleType {
	case "PLAYER":
		fpath = PATHCSVPLAYER
	case "DR":
		fpath = PATHCSVDEATH
	case "ITEM":
		fpath = PATHCSV
	}
	if fpath == "" {
		panic("TN ERROR.")
	}
	fs, err := os.Create(fpath)
	checkErr(err)
	c := csv.NewWriter(fs)
	c.WriteAll(ret)
	c.Flush()
	defer fs.Close()

}
