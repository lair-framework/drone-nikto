# drone-nikto
Imports XML from [Nikto](https://github.com/sullo/nikto) into a lair project.

## Note
The latest version of Nikto v2.1.6 contains some flaws within the XML output format that will need to be addressed before successful execution of drone-nikto. There is currently a Nikto pull request to address this issue [here](https://github.com/sullo/nikto/pull/296).

Please follow these instructions to modify existing Nikto XML files.
 * Above the first `<niktoscan>` XML tag append `<niktoscan>`
 * Append `</niktoscan>` tag to each instance of `</scandetails>` tag

The Nikto XML template can be updated by performing the following modifications:
 * <Nikto Install Dir>/templates/xml_start.tmpl append `<niktoscan>` to the end of the file.
 * <Nikto Install Dir>/templates/xml_end.tmpl append `</niktoscan>` to the end of the file.

To update Kali Linux, the xml_start.tmpl and xml_end.tmpl can be located in /usr/share/nikto/templates.

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
