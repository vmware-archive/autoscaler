# autoscaler
This is a experimental prototype of auto scaler functionality to scale the instances that are mapped to a specific target/queue/destination based on some external monitoring engine that would recommend the action to take. The monitoring aspect is outside of this autoscaler and its more of a end invoker to change the instance scales. The monitoring system should poll or monitor the end state actively (resources or queues or other systems) to come up with the final decision before handling off the scaling action to the autoscaler.

Run "go get ..." to install the dependencies and then install the autoscaler binary

# Note
Dont push this as an app to CF yet (as the cf binary is not bundled with it)
Just run the autoscaler locally with CF client utility in PATH and logged into a CF API endpoint in order to map the app to the correct org/space/app instance.

# Running the appscaler
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
]" ``` 

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
 Sample output: ``` { 'DestinationName' : 'testQ4', 'AppName' : 'hystrix-sample', 'Instances': '1'} ```

* PUT /scale/{queueName}/up/{increments}  
 Description: Scale up the mapped app instances associated with the Queue in increments

* PUT /scale/{queueName}/down/{increments}  
 Description: Scale down the mapped app instances associated with the Queue in increments

