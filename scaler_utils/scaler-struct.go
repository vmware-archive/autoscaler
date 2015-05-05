package scaler_utils

type AppDetail struct {
  Target  string `db:"target"  json:"queue"`
  AppName string `db:"appname" json:"app"`
  Org     string `db:"org"     json:"org"`
  Space   string `db:"space"   json:"space"`
  Updated int64  `db:"updated" json:"-"`
}

type AppInstanceDetail struct {
  Target   string
  AppName   string
  AppGuid   string
  Instances int
  LastFetchTime int64
}


type VcapServiceTypeDefn struct {
	Name string `json:"name"`
	Label string `json:"label"`
	Plan  string `json:"plan"`
	Tags []string `json:"tags"`
	
	Credentials struct {
		Name string `json:"name"`
		Hostname string `json:"hostname"`
		Uri string `json:"uri"`
		JdbcUrl string `json:"jbdbcUrl"`
		Port string `json:"port"`
		Password string `json:"password"`
		Username string `json:"username"`
	} `json:"credentials"`
	
}
	
// EDIT the mongo db version based on actual vcap services output 
//       if going with MongoDB
type VcapServices struct {
	MySqlServiceDefn []VcapServiceTypeDefn `json:"cleardb-n/a"`
//	MongoDbServiceDefn []VcapServiceTypeDefn `json:"mongodb-1.8`
}


/*
type Credentials struct {
	Name string `json:"name"`
	Hostname string `json:"hostname"`
	Uri string `json:"uri"`
	JdbcUrl string `json:"jbdbcUrl"`
	Port string `json:"port"`
	Password string `json:"password"`
	Username string `json:"username"`
} 
*/
