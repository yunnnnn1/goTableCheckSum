package checksum

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"goProject/flag"
	mgorm "goProject/mgorm/ExecQuerySQL"
	"hash/crc32"
	"strings"
)

func CRC32(str []string) uint32{
	aa := strings.Join(str,"")
	return crc32.ChecksumIEEE([]byte(aa))
}

func MD5(str []string) string {
	aa := strings.Join(str,"")
	c := md5.New()
	c.Write([]byte(aa))
	return hex.EncodeToString(c.Sum(nil))
}


func SHA1(str []string) string{
	aa := strings.Join(str,"")
	c:=sha1.New()
	c.Write([]byte(aa))
	return hex.EncodeToString(c.Sum(nil))
}
func Arrcmap(src ,dest []string ) []string{
	msrc := make(map[string]byte)  //按源数组建索引
	mall := make(map[string]byte)  //源+目所有元素建索引
	var set []string //交集
	//1、源数组建立map
	for _,v := range src{
		msrc[v] = 0
		mall[v] = 0
	}
	for _,v := range dest{
		l := len(mall)
		mall[v] = 1
		if l != len(mall){
			l = len(mall)
		}else {
			set = append(set,v)
		}
	}
	return set
}

func Arrcmp(src []string, dest []string) ([]string,[]string) {   //对比数据
	msrc := make(map[string]byte) //按目数组建索引
	mall := make(map[string]byte) //源+目所有元素建索引  并集
	var set []string //交集
	//1.目数组建立map
	for _, v := range dest {
		msrc[v] = 0
		mall[v] = 0
	}

	//2.源数组中，存不进去，即重复元素，所有存不进去的集合就是并集
	for _, v := range src {
		l := len(mall)
		mall[v] = 1
		if l != len(mall) { //长度变化，即可以存
			l = len(mall)
		} else { //存不了，进并集
			set = append(set, v)
		}
	}
	//3.遍历交集，在并集中找，找到就从并集中删，删完后就是补集（即并-交=所有变化的元素）
	for _, v := range set {
		delete(mall, v)
	}
	//4.此时，mall是补集，所有元素去源中找，找到就是删除的，找不到的必定能在目数组中找到，即新加的
	var added,deleted []string
	for v, _ := range mall {
		_, exist := msrc[v]
		if exist {
			deleted = append(deleted, v)
		}else {
			added = append(added,v)
		}
	}
	return added,deleted
}
func ColumnsValidation (sour,dest []byte) []string{   //校验表结构相同的表并返回
	var aa []string
	var bb []string
	var soura,desta []string
	if strings.Index(string(sour),";") != -1 {
		soura = strings.Split(strings.ToUpper(string(sour)), ";")
		soura = soura[:len(soura)-1]
		aa = soura
	}
	if strings.Index(string(dest),";") != -1 {
		desta = strings.Split(strings.ToUpper(string(dest)),";")
		desta = desta[:len(desta)-1]
	}

	if CRC32(soura) != CRC32(desta) {
		aa = Arrcmap(soura,desta)
	}

	if len(sour) !=0 {
		for i := range aa {
			bb = append(bb, strings.Split(aa[i], ";")[0])
		}
	}
	return bb
}


func ChunkValidation (a map[string]*mgorm.Connection,o *mgorm.SummaryInfo,p *flag.ConnParameter,sour,dest []byte) {
	soura := strings.Split(string(sour),"&,")
	desta := strings.Split(string(dest),"&,")
	var aa,bb []string
	if p.CheckSum == "CRC32" || p.CheckSum == "crc32"{
		if CRC32(soura) != CRC32(desta) {
			aa,bb = Arrcmp(soura,desta)
		}
	}else if p.CheckSum == "MD5" || p.CheckSum == "md5"{
		if MD5(soura) != MD5(desta) {
			aa,bb = Arrcmp(soura,desta)
		}
	}else if p.CheckSum == "SHA1" || p.CheckSum == "SHA1"{
		if SHA1(soura) != SHA1(desta) {
			aa,bb = Arrcmp(soura,desta)
		}
	}

	if bb != nil{
		DeleteSql := DestDelete(a,o,bb)
		if p.Datafix =="file"{
			fmt.Printf("Start the repair Delete SQL and write the repair SQL to /tmp/%s_%s.sql\n",o.Database,o.Tablename)
			//fmt.Printf("Start the repair Delete SQL and write the repair SQL to C:\\%s_%s.sql\n",o.Database,o.Tablename)
			SqlFile(o.Database,o.Tablename,DeleteSql)
		}else if p.Datafix =="table"{
			fmt.Printf("Start executing Delete SQL statements in the target databases %s\n",o.Database)
			a["dest"].SqlExec(DeleteSql,o)
		}
	}

	if aa != nil {
		InsertSql := DestInsert(a,o,aa)
		for _,sql := range InsertSql {
			if p.Datafix == "file" {
				fmt.Printf("Start the repair Insert SQL and write the repair SQL to /tmp/%s_%s.sql\n",o.Database,o.Tablename)
				//fmt.Printf("Start the repair Insert SQL and write the repair SQL to C:\\%s_%s.sql\n", o.Database, o.Tablename)
				SqlFile(o.Database, o.Tablename, sql)
			} else if p.Datafix == "table" {
				fmt.Printf("Start executing Insert SQL statements in the target databases %s\n", o.Database)
				a["dest"].SqlExec(sql, o)
			}
		}
	}

}
