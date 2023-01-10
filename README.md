# ev-notification

GO project to send a push notification through pushbullet when the status of a mobilityplus charging location changes.

## Running

A docker image is publicly available at https://hub.docker.com/r/yfrickx/ev-notification.
Either a stable tag can be used, which corresponds to github releases or the alpha tag can be used, which is build from the latest commit on the main branch.

To run the project you can use docker, following environment variables can be set:

| Env variable    | Required                      | use                                                                                                                                                                                                                  |
|-----------------|-------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| PB_KEY          | Yes                           | Pushbullet api key, can be retrieved through the website                                                                                                                                                             |
| DEVICE_ID       | Yes                           | Pushbullet device id to which the notifications should be sent (check https://docs.pushbullet.com/#list-devices to see how to fetch it)                                                                              |
| MBP_LOCATION    | No, if MBP_CONFIG_FILE is set | Location ID of the MobilityPlus charging station (go to https://www.mobilityplus.be/en/map and click on a specific location, the id can be found through the network calls when the details of a station are opened) |
| MBP_CONFIG_FILE | No, if MBP_LOCATION is set    | Yaml config file to setup multiple locations, example can be found below                                                                                                                                             |

Example config file:
```yaml
locationGroups:
  - groupName: "group1"
    locations:
      - "1"
  - groupName: "group2"
    locations:
      - "2"
      - "3"
      - "4"
```