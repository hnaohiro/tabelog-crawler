package crawler

import (
  "fmt"
  "strconv"
  "strings"
  "errors"
  "reflect"
  "encoding/xml"
  "io/ioutil"
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
)

type Tabelog struct {
  apikey string
  httpclient *HttpClient
  db *sql.DB
}

func NewTabelog() (*Tabelog, error) {
  apikey, err := readAPIKey()
  if err != nil {
    return nil, err
  }

  db, err := initDB()
  if err != nil {
    return nil, err
  }

  tabelog := new(Tabelog)
  tabelog.apikey = apikey
  tabelog.httpclient = NewHttpClient()
  tabelog.db = db

  return tabelog, nil
}

func readAPIKey() (string, error) {
  b, err := ioutil.ReadFile("./config/apikey")
  if err != nil {
    return "", err
  }
  return strings.TrimSpace(string(b)), nil
}

func initDB() (*sql.DB, error) {
  db, err := sql.Open("sqlite3", "./db/tabelog.db")
  if err != nil {
    return nil, err
  }

  fields, _ := GetStructData(new(Restaurant))
  sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS restaurants(%s)", strings.Join(fields, ","))
  if _, err = db.Exec(sql); err != nil {
    return nil, err
  }

  fields, _ = GetStructData(new(Review))
  sql = fmt.Sprintf("CREATE TABLE IF NOT EXISTS reviews(%s)", strings.Join(fields, ","))
  if _, err := db.Exec(sql); err != nil {
    return nil, err
  }

  return db, nil
}

func GetStructData(structType interface{}) (fields []string, values []interface{}) {
  elem := reflect.ValueOf(structType).Elem()
  size := elem.NumField()

  fields = make([]string, size)
  values = make([]interface{}, size)

  for i := 0; i < size; i++ {
    fields[i] = elem.Type().Field(i).Name
    values[i] = elem.Field(i).Interface()
  }

  return fields, values
}

func (tabelog *Tabelog) CloseDB() {
  tabelog.db.Close()
}

func (tabelog *Tabelog) Get(station string, page int) (*RestaurantInfo, error) {
  url := "http://api.tabelog.com/Ver2.1/RestaurantSearch/"
  url += "?Key=" + tabelog.apikey
  url += "&PageNum=" + strconv.Itoa(page)
  url += "&ResultSet=large"
  url += "&Station=" + urlencode(station)

  res, err := tabelog.httpclient.Get(url)
  if err != nil {
    return nil, err
  }

  var info RestaurantInfo
  if err = xml.Unmarshal(res, &info); err != nil {
    return nil, err
  }

  if info.NumOfResult ==  0 {
    return nil, parseError(res)
  }

  return &info, nil
}

func urlencode(s string) (result string){
  for _, c := range(s) {
    if c <= 0x7f { // single byte 
      result += fmt.Sprintf("%%%X", c)
    } else if c > 0x1fffff {// quaternary byte
      result += fmt.Sprintf("%%%X%%%X%%%X%%%X",
        0xf0 + ((c & 0x1c0000) >> 18),
        0x80 + ((c & 0x3f000) >> 12),
        0x80 + ((c & 0xfc0) >> 6),
        0x80 + (c & 0x3f),
      )
    } else if c > 0x7ff { // triple byte
      result += fmt.Sprintf("%%%X%%%X%%%X",
        0xe0 + ((c & 0xf000) >> 12),
        0x80 + ((c & 0xfc0) >> 6),
        0x80 + (c & 0x3f),
      )
    } else { // double byte
      result += fmt.Sprintf("%%%X%%%X",
        0xc0 + ((c & 0x7c0) >> 6),
        0x80 + (c & 0x3f),
      )
    }
  }

  return result
}

func (tabelog *Tabelog) GetReviews(Rcd int) (*ReviewInfo, error){
  url := "http://api.tabelog.com/Ver1/ReviewSearch/"
  url += "?Key=" + tabelog.apikey
  url += "&Rcd=" + strconv.Itoa(Rcd)

  res, err := tabelog.httpclient.Get(url)
  if err != nil {
    return nil, err
  }

  var info ReviewInfo
  if xml.Unmarshal(res, &info); err != nil {
    return nil, err
  }

  if info.NumOfResult ==  0 {
    return nil, parseError(res)
  }

  return &info, nil
}

func parseError(res []byte) error {
  type ApiError struct {
    Message string
  }

  var apiError ApiError
  if err := xml.Unmarshal(res, &apiError); err != nil {
    return err
  }
  return errors.New(apiError.Message)
}

func (tabelog *Tabelog) Save(tableName string, data interface{}) error {
  fields, values := GetStructData(data)
  size := len(fields)

  sql := fmt.Sprintf("INSERT INTO %s(%s) values(%s)", tableName, strings.Join(fields, ","), strings.Repeat("?,", size)[:size*2-1])
  stmt, err := tabelog.db.Prepare(sql)
  if err != nil {
    fmt.Println(sql)
    return err
  }

  _, err = stmt.Exec(values...)
  return err
}

type Restaurant struct {
  Rcd int                 // レストランID
  RestaurantName string   // レストラン名
  TabelogUrl string       // レストラン詳細ページ（PC）のURL
  TabelogMobileUrl string // レストラン詳細ページ(モバイル)のURL
  TotalScore string       // 点数（総合評価）
  TasteScore string       // 料理・味の点数
  ServiceScore string     // サービスの点数
  MoodScore string        // 雰囲気の点数
  Situation string        // シチュエーション
  DinnerPrice string      // 価格（夜）
  LunchPrice string       // 価格（昼）
  Category string         // ジャンル名
  Station string          // 最寄り駅
  Address string          // 住所
  Tel string              // 電話番号
  BusinessHours string    // 営業時間
  Holiday string          // 休日
  Latitude string         // 緯度
  Longitude string        // 経度
}

type RestaurantInfo struct {
  NumOfResult int
  Item []Restaurant
}

type Review struct {
  NickName string       // 投稿したユーザーのニックネーム
  VisitDate string      // レストラン訪問日
  ReviewDate string     // 口コミ投稿日
  UseType string        // 口コミ対象（夜のみ、昼のみ、夜・昼両方）
  Situations string     // オススメシチュエーション（友人・同僚と、デート、接待、宴会、家族・子供と、一人で ）
  TotalScore string     // 点数（総合評価）
  TasteScore string     // 料理・味の点数
  ServiceScore string   // サービスの点数
  MoodScore string      // 雰囲気の点数
  DinnerPrice string    // 使った金額/1人当り(夜)
  LunchPrice string     // 使った金額/1人当り(昼)
  Title string          // 口コミのタイトル
  Comment string        // 口コミのコメント（文頭より99文字まで）
  PcSiteUrl string      // 口コミページ（PC）のURL
  MobileSiteUrl string  // 口コミページ（モバイル）のURL
}

type ReviewInfo struct {
  NumOfResult int
  Item []Review
}
