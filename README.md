# Map

## Initial Setup

### Clone Map, build and install it using GO

```git clone git@github.com:kramanathan01/urlshort.git && cd urlshort/map
$ go install .
```

### Create .map.json file in the home directory

Setup a JSON file with short path and matching URL. See example below.
```
echo "
[{
    "path": "/nyt",
    "url": "https://www.nytimes.com"
  },
  {
    "path": "/wp",
    "url": "https://www.washingtonpost.com"
  }
]" > ~/.map.json
```

### Enjoy browsing

You can go to your favorite browser and use the shortcuts by typing in:
```
http://localhost:8080/nyt
```

## Simplifying access

A few additional steps can save time when browsing

### Setup DNS

Map runs on localhost. For easier access, setup a name in /etc/hosts

```sudo echo "map      127.0.0.1" >> /etc/hosts```

### Setup port forwarding

Map uses port 8080. By forwarding port 80 to 8080, you can save typing the port everytime.
This is a bit involved in MacOS, Yosemite onwards.

1. Enable port forwarding

```
sudo sysctl net.inet.ip.forwarding=1
```

2. Create an anchor file

```
sudo echo “rdr pass on lo0 inet proto tcp from any to 127.0.0.1 port 80 -> 127.0.0.1 port 8080” > /etc/pf.anchors/map
```

3. Add the following lines to /etc/pf.conf

```
rdr-anchor “map”
load anchor "map" from "/etc/pf.anchors/map"
```

### Enjoy the shortcut

Now you can access your favorite shortcuts by typing in:
```
http://map/nyt
```

