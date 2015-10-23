# drone-nikto
Imports XML from [Nikto](https://github.com/sullo/nikto) into a lair project.

## Note
The latest version of Nikto v2.1.6 contains some flaws within the XML output format that will need to be addressed before successful execution of drone-nikto. Please follow the following instructions for updates to be applied to the Nikto XML file.
 * Above the first `<niktoscan>` XML tag append `<niktoscan>`
 * Append `</niktoscan>` tag to each instance of `</scandetails>` tag

## Install
Download a compiled binary for supported operating systems from [here](https://github.com/lair-framework/drone-nikto/releases/latest).

```
$ mv drone-nikto* drone-nikto
$ ./drone-nikto -h
```

## Build from source
```
$ go get github.com/lair-framework/drone-nikto
```
