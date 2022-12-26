# ev-notification

GO project to send a push notification through pushbullet when the status of a mobilityplus charging location changes.

## Running

To run the project you can use docker, three environment variables should be set:
| Env variable | use                                                                                                                                                                                                                  |
|--------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| PB_KEY       | Pushbullet api key, can be retrieved through the website                                                                                                                                                             |
| DEVICE_ID    | Pushbullet device id to which the notifications should be sent (check https://docs.pushbullet.com/#list-devices to see how to fetch it)                                                                              |
| MBP_LOCATION | Location ID of the MobilityPlus charging station (go to https://www.mobilityplus.be/en/map and click on a specific location, the id can be found through the network calls when the details of a station are opened) |
