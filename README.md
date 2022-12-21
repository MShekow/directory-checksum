# Directory Checksum
![Coverage](https://img.shields.io/badge/Coverage-96.6%25-brightgreen)
![CI](https://github.com/MShekow/directory-checksum/actions/workflows/ci.yml/badge.svg)
![CD](https://github.com/MShekow/directory-checksum/actions/workflows/cd.yml/badge.svg)

This tool recursively computes the checksum of the contents of a directory, and prints the checksums up to a depth you
can specify.

_Symbolic links_ are ignored and are not followed.

This tool is a _proper_ (as in: actually working) alternative to tools that ignore _empty_ directories, such
as [md5deep](https://md5deep.sourceforge.net/), or chaining UNIX such as `find` and `md5sum` (see
e.g. [here](https://unix.stackexchange.com/a/35834)).

See here TODO for background information why I wrote this tool. Its primary use case is to debug layer caching issues
with Docker, BuildKit, Buildah or other OCI image builder tools, which rebuild a layer with a `COPY` or `ADD` statement,
even though you think it should not have been rebuilt, because none of the files have changed.

## Example output

Running `directory-checksum` for the current working directory, printing up to 4 levels.

```shell
$ directory-checksum --max-depth=4 .

7c47daae101786a01cccf330884ba7c7a3ecb91e D .
a6bb67a68b83ce1daabc176df5591f03bd9c2078 D .idea
7328aae6b9c0f5d22065fb856dd373ab4b999f3b D .idea\inspectionProfiles
e30fb4cb0ba9888edb9f327b0a8e391bd6df2f97 F .idea\inspectionProfiles\Project_Default.xml
f407904694bc6b866c2c0c732828ef8478450583 F .idea\.gitignore
ce2de3130718fd2af4d75e902ed17e43f4d4a7a3 F .idea\golang-exp1.iml
46ba4606e52739668fc028814845a40235a9675c F .idea\modules.xml
89514dec2f816a283d6616bb9bbf686e199f2a3b F .idea\workspace.xml
56d5471f798fd45fcef20db937e4c2ed26aea0d2 D directory_checksum
8b6d8131a485c39454dbd5e2d9ff0be6a0151c6f F directory_checksum\checksum_utils.go        
fdc1a074e561ccde414db93bcde7512f968e4fc2 F directory_checksum\checksum_utils_test.go   
8e7f1326626b0850e485d3f7a80a5f5aed214eb3 F directory_checksum\dict_utils.go
f7ad8b95c6b22ee16b6cb56374b23f6107005159 F directory_checksum\dict_utils_test.go       
0497158638db643c2c6b1ae732b5290f59fb28e2 F directory_checksum\directory.go
2dfe102720acd43ba58333de36704e764bc77341 F directory_checksum\fs_scanner.go
e04c3bbede11f5d49309619396afd35eab361d59 F directory_checksum\fs_scanner_test.go       
0740118c1fdc3944256368be553b7e074a573fb3 F .gitignore
0e8c89b6a2f9d6067805f4f9b3b30d53c9cae8e9 F README.md
9a06b436f851ccd8756d10566109451ba429208b F go.mod
04cf991671c86b0226ac06bcfcdd70d3aeb33687 F go.sum
d1fbe91a2253ba9de3ced8f95ea7a8b30d436727 F main.go
```

Explanation of columns:

- First column shows **SHA-1** checksum of files and directories.
    - For _files_, only the binary file content is used for computing the checksum, all other meta-data (e.g. owner) are
      ignored
    - For _directories_, the checksum is the **SHA-1** checksum of a long string that represents the listing of the
      _immediate_ children, only considering their _names_ and checksums
- Second column: `D`=_directory_, `F`=_file_
- Third column: the path _relative_ to the scanned directory's path

Note: the first line always shows the checksum of the scanned directory itself.

## Use case: debug image build caching issues

A common problem is that commands such as `docker build ...` rebuild an image layer (for an `ADD` or `COPY` statement in
a `Dockerfile`) even though it should _not_ have done so: from your point of view, _none_ of the files have changed, and
therefore the image layer cache should have been used, instead of rebuilding the layer.

Thanks to _Directory Checksum_, you can now easily debug the problem, by modifying your `Dockerfile` as follows:

- Add the `directory-checksum` binary to your image, in some layer that _precedes_ the problematic `COPY`/`ADD` layer:
    - You may want to use a statement such as `ADD <URL> /usr/local/bin/directory-checksum` followed
      by `RUN chmod +x /usr/local/bin/directory-checksum`, to avoid that you first have install `curl`. Replace `<URL>`
      with the appropriate [binary release](https://github.com/MShekow/directory-checksum/releases) of _Directory
      Checksum_.
    - If you use _BuildKit_, _Buildah_ or some other build engine that supports `ADD --chmod`, you can
      use `ADD --chmod=755 <URL> /usr/local/bin/directory-checksum`
- Add a statement such as `RUN directory-checksum --max-depth 2 .` to print the checksums of the directory.
    - Replace `2` with any other depth-level, if desired. A too large number will produce too much output, a too small
      number may provide too few details (especially when something changed in the deeper levels of the folder
      hierarchy).
    - Replace `.` with any other path that is either a _relative_ path to your `WORKDIR`, or an _absolute_ path.

**Note:** we run `directory-checksum` _inside the build container_ (not on the _host_) so that the filtering applied by
the  `.dockerignore` file is already accounted for.

Now, to find out why the image layer cache has not been used for your `COPY`/`ADD` layer, you simple compare the output
of `directory-checksum` between two `docker build` executions. Once you determined the files or directories that have
changed, you can tweak your `.dockerignore` file accordingly (or file a bug with your container build engine if your
files _really_ have not changed).

## Building and testing

This is a simple CLI application implemented in _Go_, thus I assume that you are familiar with how to build Go
applications (`go build`) or how to run tests (`go test ./... -v`).

