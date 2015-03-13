cask configuration (meta.json)
------------------------------

meta.json configuration format is a JSON file with the following parameters

- runtime - runtime to use when building the image
- name - runtime to use when building the image
- config - additional configuration for the container
- data - arbitrary  data for the image
- build.images - list of images to include in the image
- build.exclude - ilst of files to remove according to a glob match

example

      {
         "runtime": "ubuntu12",
         "name": "test",
         "config" : {
            "lxc.init_cmd": "/usr/bin/runsvdir -P /etc/service"
         },
         "build" : {
            "images" : [
               "runit"
            ],
            "exclude" : [
               "/var/lib/dpkg/*",
               "/var/log/*",
               "/var/cache/apt/*",
               "/var/lib/apt/*",
               "/usr/share/man/man*",
               "/usr/share/doc*"
            ]
         }
      }

