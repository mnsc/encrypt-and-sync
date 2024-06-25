# Encrypt and sync

This is a small cli tool for syncing a photo collection to OneDrive. 

The idea is to utilize OneDrive as a backend in as simple way as possible, without talking to any API:s, just copy media files (raw photos and videos) located on my hard drive to a mirrored folder structure insided my OneDrive folder, while at the same time encrypting them with a 256 bit AES key (GCM). The file name includes the hash of the source file and an additional ".encr". Files that I don't consider sensitive, like metadata files (xmp, ini, txt), are not encrypted but just copied over to the OneDrive folder.

```

realphotos/                  backuphotos/                                 
├── 2023/                    ├── 2023/              
│   ├── January/             │   ├── January/           
│   │   ├── photo1.cr2       │   │   ├── photo1.cr2.b409...d5c31.encr
│   │   ├── photo2.jpg       │   │   ├── photo2.jpg.c8a6...6302c.encr
│   │   └── video1.mov       │   │   └── video1.mov.52c1...e0b7e.encr
│   └── February/            │   └── February/           
│       ├── photo3.jpg       │       ├── photo3.jpg.b928...cdc41.encr           
│       ├── photo3.xmp       │       ├── photo3.xmp
│       └── video2.mov       │       └── video2.mov.3087...15a42.encr
└── 2022/                    └── 2022/              
    └── December/                └── December/           
        ├── photo4.jpg               ├── photo4.jpg.fd9b...ca0e9.encr
        ├── notes.txt                ├── notes.txt
        └── video3.mp4               └── video3.mp4.20bde...a6ea.encr           
```

## OMGAI!

It was also an excercise in using Cursor's plus plan to be able to create a niche program completely from scratch with the help of an LLM. In the end it worked out pretty well and I haven't written many of these code lines "myself" and it felt like a natural way to work, like pair programming with someone. Ableit someone that sometimes just doesn't listen to me and produces unexpected/wrong code, mainly when asking it to do refactorings or other code changes that touches multiple files. But in the end it feels like the future and I invested about 8 hours in this and I'm happy with the result. 


## Usage

### Syncing files to OneDrive

```
go run . -sourcefolder "C:\path\to\your\photos" -onedrivefolder "C:\path\to\your\onedrive\photos"
```

The program looks for an encryption key in the environment variable ENCRYPTION_KEY and if that is not found it will prompt for it on program start.

### Restoring

To restore files from OneDrive to your hard drive:

```
go run . -onedrivefolder "C:\path\to\your\onedrive\photos" -restore "C:\path\to\your\restored\photos"
```


### Notes

Since this tool is expected to be run continously as I add new photos to my hard drive it will add a "keyfile" to the root of the onedrive folder. This is not a keyfile but a small json encrypted with the given key. On the next run it will try to decrypt this file with the key given and read the json. This is to ensure that all files in the folder are encrypted with the same key. 