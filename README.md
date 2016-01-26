# Generatorbeat

 Welcome to Generatorbeat.

Ensure that this folder is at the following location:
`${GOPATH}/github.com/urso/`

## To get running with Generatorbeat, run the following commands:

```
glide init
glide update --no-recursive
make update
make
```


## To generate etc/generatorbeat.template.json and etc/generatorbeat.asciidoc

```
make generate
```

## To run Generatorbeat with debugging output enabled, run:

```
./generatorbeat -c etc/generatorbeat.yml -e -d "*"
```

## To test Generatorbeat, run the following commands:

```
make testsuite
```

alternatively:
```
make unit-tests
make system-tests
make integration-tests
make coverage-report
```


The test coverage is reported in the folder `./build/coverage/`

## To clean  Generatorbeat source code, run the following commands:

```
make fmt
make simplify
```

## To package Generatorbeat for all platforms, run the following commands:

```
cd packer
make
```


## To push Generatorbeat in the git repository, run the following commands:

```
git init
git add .
git commit
git remote set-url origin https://github.com/urso//generatorbeat
git push origin master
```

## To clone Generatorbeat from the git repository, run the following commands:

```
mkdir -p ${GOPATH}/github.com/urso/
cd ${GOPATH}/github.com/urso/
git clone https://github.com/urso//generatorbeat
```


## For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).
