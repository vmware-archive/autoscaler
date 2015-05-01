package scaler_utils

type AppDetail struct {
  TargetQ string `json:"queue"`
  AppName string `json:"app"`
  Org     string `json:"org"`
  Space   string `json:"space"`
}

type AppInstanceDetail struct {
  TargetQ   string
  AppName   string
  AppGuid   string
  Instances int
}
