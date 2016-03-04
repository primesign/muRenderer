# muRenderer

muRenderer is a small REST service for rendering pdfs and build around [mupdf](http://mupdf.com/).

## Building an executable

Building muRenderer is a two step process:

1.) build mupdf 
2.) build muRenderer itself

To build mupdf for Linux and Windows on Debian/Ubuntu install the following packages:

``sudo apt-get install build-essential gcc-mingw-w64``

Fetch the mupdf source code with:

``git submodule update --init --recursive``

Call the build script:

``./build.sh``

The output are two executables:

1. ``renderservice.exe`` for Windows
2. ``renderservice`` for Linux

## Usage

Use ``./renderservice -h`` to see all available options.

### REST API
#### GET
    http://host:port/renderservice/health
    http://host:port/renderservice/{documentId}/{pageNr}/pageinfo
    http://host:port/renderservice/{documentId}/{pageNr}/numpages
    http://host:port/renderservice/{documentId}/{pageNr}?z={zoomfactor}
    http://host:port/renderservice/{documentId}/{pageNr}?w={width}&h={height}

#### POST`
    http://host:port/renderservice/{documentId}/{pageNr}

#### DELETE
    http://host:port/renderservice/{documentId}


## Contributing
Contributions are very welcome.

1. Fork the project
2. Create a feature branch: `git checkout -b my-new-feature`
3. Commit changes: `git commit -am 'Add my feature'`
4. Push to the branch: `git push -u origin my-new-feature`
5. Submit a pull request
