# autoscaler is no longer actively maintained by VMware.

# autoscaler
This is a experimental prototype of auto scaler functionality to scale the instances that are mapped to a specific target/queue/destination based on some external monitoring engine that would recommend the action to take. 

The monitoring aspect is outside of this autoscaler and its more of a end invoker to change the instance based on request to scale up or down. The monitoring system should poll or monitor the end state actively (resources or queue depth or other constraints) to come up with the decision to scale bnd handling off the scaling action to the autoscaler.

Run "go get ..." to install the dependencies and then run "go install" to install the autoscaler binary to run locally.
For testing on CF, just following the section below.

# Running on CF or Locally
There are two ways to test:
1) Push this as an application to CF 
2) Just run the autoscaler locally with CF 

The autoscaler application would require three environment variables to set for the Application to scale the managed app instances. Check the Credentials page of the Pivotal Elastic Runtime to get the following property (under UAA -> client section):
```
Sample entry inside the Ops Mgr -> Elastic Runtime Tile -> Credentials tab -> Search for Autoscaling

Autoscale Client Credentials  autoscaling_service / fe20d9a573c60b220a02
```
.
* UAA_CLIENT_ID
  This should be set to `autoscaling_service`.
  If its not possible or autoscaling_service client does not exist, please see the check on creating the uaac client tokens
* UAA_CLIENT_SECRET
  This should be set to the generated password or associated password for the corresponding client `autoscaling_service`
* DOMAIN
  This should be set to the CF's system domain path (without protocol, ex: `10.244.0.34.xip.io` if running against some xip addr or `system.cf-app.com`)

If running the autoscaler app locally (not inside CF), do exports of the variable
```
  export UAA_CLIENT_ID=autoscaling_service
  export UAA_CLIENT_SECRET=_FILL_ME_
  export DOMAIN=_FILL_ME_               #Example: 10.244.0.34.xip.io
```
If running the autoscaler app by pushing to CF, use set-env to set the variables before restaging or specify in manifest.yml
```
  cf set-env autoscaler UAA_CLIENT_ID autoscaling_service
  cf set-env autoscaler UAA_CLIENT_SECRET _FILL_ME_
  cf set-env autoscaler DOMAIN            _FILL_ME_   #Example: 10.244.0.34.xip.io
```

# Persistence
Its possible to allow mysql service binding to be used for persisting the app state to survive restarts/crashes. Uncomment the services section inside the manifest and specify the service instance name. Any new registration of apps against targets would be persisted to DB and reloaded on startup.

# Running the appscaler locally
The appscaler uses default port of 8080 (which can be overridden via an Env variable PORT that can specify a different port).
Following are the REST style endpoints exposed to take action:

# Exposed Methods:  
* POST /register  
 Description: Register or update dynamically a single destination/queue with an app (with org and space details) with the Autoscaler.
 JSON Input: ```  ' { "queue" : "RabbitQ", "app" : "rabbitmq-consumer", "org" : "Logistics", "space" : "preprod" } ' ```

* POST /load  
 Description: Bulk register of multiple destination/queue app sets (with org and space details) with the Autoscaler.  
 JSON Input 
 ```
 "[ { "queue" : "testQ1", "app" : "eureka-service", "org" : "platform", "space" : "josh" },
    { "queue" : "testQ2", "app" : "cf-bg-demo-blue", "org" : "platform", "space" : "mine" }, 
    { "queue" : "testQ5", "app" : "test-bp", "org" : "platform", "space" : "test" }
 ]"
 ``` 

* DELETE /register/{queueName}  
 Description: Remove the mapping of the destination/queue and app from the Autoscaler.    

* GET /details  
 Description: Dump all data about all registered destinations/queues and associated app instance details    
 JSON Input:
 ```
 "[{ 'DestinationName' : 'testQ2', 'AppName' : 'cf-bg-demo-blue', 'Instances': '1'}
   { 'DestinationName' : 'testQ3', 'AppName' : 'my-cgm-green', 'Instances': '1'}
   { 'DestinationName' : 'testQ4', 'AppName' : 'hystrix-sample', 'Instances': '1'}
   { 'DestinationName' : 'testQ5', 'AppName' : 'wls-test-bp', 'Instances': '1'}
   { 'DestinationName' : 'testQ1', 'AppName' : 'eureka-service', 'Instances': '1'}]"
  ```

* GET /details/{queueName}  
 Description: Dump data about a registered destination/queue and associated app instance details    
 Sample output: 
 ``` { 'DestinationName' : 'testQ4', 'AppName' : 'hystrix-sample', 'Instances': '1'} ```

* PUT /scale/{queueName}/up/{increments}  
 Description: Scale up the mapped app instances associated with the Queue in increments

* PUT /scale/{queueName}/down/{increments}  
 Description: Scale down the mapped app instances associated with the Queue in increments

# To run on CF

Autoscaler needs a UAA Client created ahead of time that has the requisite privileges to change the instance count of any app managed by it.
If there is already autoscaling tile installed (part of PCF install), then the same set of client credentials can be used.

# UAA Client Token for managing Autoscaling
Steps to create the UAA Client with right permissions:
```
# setup UAAC, the command line for UAA
gem install cf-uaac
uaac target uaa.<your_domain>

# get a token that is capable of creating your new UAA client
uaac token client get admin --scope "clients.read,clients.write"
# the admin client secret (client! not admin scim user) is in the CF deployment manifest under uaa properties

uaac client add autoscaler --scope "openid,cloud_controller.permissions,cloud_controller.read,cloud_controller.write" --authorized_grant_types "client_credentials,authorization_code" --authorities "cloud_controller.write,cloud_controller.read,cloud_controller.admin notifications.write critical_notifications.write emails.write" --access_token_validity 3600 --autoapprove true
# use whatever secret you want
```
You now have a UAA client (UAA_CLIENT_ID environment variable) called *autoscaler* with whatever secret you just chose above. 
You need to configure the autoscaler app to know the client secret using the UAA_CLIENT_SECRET environment variable

