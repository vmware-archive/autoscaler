package cf_utils

type ErrorResponse struct {
      Code int `json:"code"`
      Description string `json:"description"`
      Error_code string `json:"error_code"`
}

type GenericResponse struct {
  Metadata struct {
      Guid string `json:"guid"`
      Url string `json:"url"`
   } `json:"metadata"`
  Entity struct {
      Name string `json:"name"`
   } `json:"entity"`
}

type QueryResponse struct {
   Resources []GenericResponse `struct:"resources"`
}

/*
type GenericUrl struct {
      Host string `json:"host"`
}
*/

type GenericUri struct {
      Uri string `json:"uri"`
}

type GenericUsage struct {
      Time string `json:"time"`
      Cpu float64 `json:"cpu"`
      Memory int `json:"mem"`
      Disk int `json:"disk"`
}

type GenericDomain struct {
      Name string `json:"name"`
}

type GenericRoute struct {
      Host string `json:"host"`
      Domain GenericDomain `struct:"domain"`
}

type AppDetail struct {
      Guid string `json:"guid"`
      Urls []string `json:"urls"`
      Routes []GenericRoute `json:"routes"`
      Services []string `json:"service_names"`
      Running int `json:"running_instances"`
      Name string `json:"name"`
      Detected_buildpack string `json:"detected_buildpack"`
      State string `json:"state"`
      Memory int `json:"memory"`
      Instances int `json:"instances"`
      Disk int `json:"disk_quota"`
      Environment_json string `json:"environment_json"`
      Command string `json:"command"`
      Detected_start_command string `json:"detected_start_command"`
}

type ServicePlanProviderDetail struct {
      Label string `json:"label"`
      Provider string `json:"provider"`
}

type ServicePlanDetail struct {
      Guid string `json:"guid"`
      Name string `json:"name"`
      ServiceProvider ServicePlanProviderDetail `json:"service"`
}

type ServiceDetail struct {
      Guid string `json:"guid"`
      Name string `json:"name"`
      Bound_App_Count int `json:"bound_app_count"`
      Plan ServicePlanDetail `json:"service_plan"`
}

type AppStat struct {
      Name string `json:"name"`
      Uris []string `json:"uris"`
      Host string `json:"host"`
      Port int `json:"port"`
      Uptime int `json:"uptime"`
      Memory_Quota int `json:"mem_quota"`
      Disk_Quota int `json:"disk_quota"`
      Usage GenericUsage `json:"usage"`
}

type AppStatsResponse struct {
      State string `json:"state"`
      Stat AppStat `json:"stats"`
}

type SpaceSummaryResponse struct {
      Guid string `json:"guid"`
      Name string `json:"name"`
      Apps []AppDetail `struct:"apps"`
      Services []ServiceDetail `struct:"services"`
}

type SpaceResponse struct {
  Guid string `json:"guid"`
  Name string `json:"name"`
  Metadata struct {
      Guid string `json:"guid"`
      Url string `json:"url"`
   } `json:"metadata"`
  Entity struct {
      Name string `json:"name"`
   } `json:"entity"`
}

type AppSummaryResponse struct {
  Metadata struct {
      Guid string `json:"guid"`
      Url string `json:"url"`
   } `json:"metadata"`
  Entity struct {
      Name string `json:"name"`
      //Urls []GenericUrl `json:"urls"`
      //Routes []GenericRoute `json:"routes"`
      //Running int `json:"running_instances"`
      Buildpack string `json:"buildpack"`
      Detected_buildpack string `json:"detected_buildpack"`
      Environment_json string `json:"environment_json"`
      Memory int `json:"memory"`
      Instances int `json:"instances"`
      Disk int `json:"disk_quota"`
      State string `json:"state"`
      Command string `json:"command"`
      Detected_start_command string `json:"detected_start_command"`
   } `json:"entity"`
}


type SpaceAppResponse struct {
  Metadata struct {
      Guid string `json:"guid"`
      Url string `json:"url"`
   } `json:"metadata"`
  Entity struct {
      Name string `json:"name"`
	  Instances int `json:"instances"`
   } `json:"entity"`
}

type SpaceAppQueryResponse struct {
   Resources []SpaceAppResponse `struct:"resources"`
}


