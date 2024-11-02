## Commenteer

Commenteer is an app that allows you to combine an image and some comments from a reddit post into a single image.

### Getting Started

This is a server side rendered web app written in go using the built in http server and templates.

Docker is primarily used to run/deploy the application and uses the following containers:
* Commenteer app
  * the go app server
* imgproxy
  * used to process images for the app
* postgres
  * db
* traefik (prod only)
  * reverse proxy
* watchtower (prod only)
  * updates app after a new docker image is created
* pgadmin (dev only)
  * db viewer


To run the application you can start it using the following command:

``` docker compose --env-file=/run/secrets/.env.local up -d ```

this assumes you have docker installed and running, and your env file exists in /run/secrets/.env.local (C://run/secrets/.env.local on windows)
* see .env.sample for environment vars required for running the app
* this env file will provide values to both the docker compose file and the actual go application
* Currently this app is linked directly to an app created on reddit (reddit.com/prefs/apps).  I haven't tested this yet but maybe another reddit app can be created there and use this app?

if you want to live update the go app during development you can use the live reloading server air (https://github.com/air-verse/air)

to do this you need to stop the commenteer app docker container and from the base directory run the command ```air```

### Basic design

![design diagram](design/commenteer-design.png)