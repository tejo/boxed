## Boxed

A Dropbox based blog engine

This project started as a blog as a service platform, but it quickly ended up to be my
[personal blog](http://blog.parmi.it/). It does its job, it's far away to be pretty and polished but you maybe find it
useful. It allows you to manage your blog in markdown format from your dropbox
folder.

You can run it on your machine without installing anything, It's all bundled in the executable (html and css too), and it saves the data in a [bolt](https://github.com/boltdb/bolt) database. It's performance wise too, my personal blog is served by a raspberry pi behind my crappy DSL router.

### Try it

If you want to try it without compile it by yourself you can grab the executable
from the [aplha releases](https://github.com/tejo/boxed/releases/tag/v0.1-alpha)
page

You have to create a [dropbox app](https://www.dropbox.com/developers/apps) and
select the following:


- Dropbox API app
- Files and datastores
- Yes My app only needs access to files it creates.
- webhook url: http://yoursitehost.com/webhook


then you need to modify the ```.env.sample``` file accordingly and in your terminal:

```
# set env variables
source .env.sample

# link local app to dropbox (follow the instructions)
./boxed --oauth

# put a markdown file in the published folder in your dropbox directory and publish it with:
./boxed --refresh

#run the server with
./boxed

# visit localhost:8080 to see the result 
```


if you have correctly set the webhook path you don't need to refresh, it will be published when the article will been synchronized to your dropbox.


### Articles metadata

Boxed will try its best to figure out the publication date and title from the markdown article. If this is not enough for you, you can specify some metadata with a markdown/json comment like this:

```
<!--{
		"created-at": "2013-11-11",
		"permalink": "a-brand-new-blog",
		"title": "A brand new blog"
}-->

```

### Images

Boxed supports images out of the box, you have to put them in the ```images``` folder, and then reference them like: ```![cool image](../images/image.jpeg")```


### Template customization

Boxed comes with the excellent default [hyde](http://hyde.getpoole.com/) template. If you want to change it, like i did it for my [blog](http://boxed.parmi.it/), you have to be able to compile go code, then you have to change the template/css files and then use [go rice](https://github.com/GeertJohan/go.rice) to bundle them back in the executable. 

### Tests

You can run boxed tests with:

```
go test ./...
```
