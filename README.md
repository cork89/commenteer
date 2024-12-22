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
* Currently this app is linked directly to an app created on reddit (https://reddit.com/prefs/apps).
  * two apps are currently being used one for prod and one for dev

if you want to live update the go app during development you can use the live reloading server air (https://github.com/air-verse/air)

to do this you need to stop the commenteer app docker container and from the base directory run the command ```air```

### Basic design

![design diagram](design/commenteer-design.avif)

#### Basic workflow
* User inputs a url link to a reddit comment of an image post
* Commenteer does validation on the link, then calls the reddit api to get the json of the post and local comment tree
* Edit page is loaded with the image being locally downloaded to imgproxy and resized + other processing and the comments.
  * This initial info of the post and comments is saved to the db
* User presses "Publish" button on the edit page, the post and comment are saved to an image and sent to the r2 bucket, page is redirected to the view page for that newly created image

### Additional Dev Setup

#### Tailwind

* this project is currently using tailwind for styling, utilizing tailwind cli 
  * instructions for setup can be found here: https://tailwindcss.com/docs/installation
  * this command can be run to build the css 
    * ```npx tailwindcss -i ./static/input.css -o ./static/style.css --minify```