const fs = require('fs');
const path = require('path');
const axios = require('axios');
const showdown = require('showdown');
const pathData = require('./docPath.json');


const CWD = process.cwd();

const saveFile = (filePathDest, content) => {
    fs.writeFile(filePathDest, content, err => {
        if (err) {
            console.error(err);
        }
    });
};
const convertFile = (data) => {
    const converter = new showdown.Converter();
    return converter.makeHtml(data); 
}
const convertLocalFile = (filePathSrc, filePathDest) => {
    fs.readFile(CWD + filePathSrc, 'utf8', (err, data) => {
        if (err) {
            console.error(err);
            return;
        }
        saveFile(CWD + filePathDest, convertFile(data));
    });
}

const convertLocalFiles = (pathInfo) => {
    const srcDir = pathInfo.source;
    const destDir = pathInfo.destination;

    fs.readdir(CWD + srcDir, (err, files) => {
        if (err) {
            console.error(err);
            return;
        }
        files.forEach(file => {
            convertLocalFile(
                srcDir + file,
                destDir + file.replace('.md', '.html')
            );
        });
    });
}
const getDestination = (pathInfo, fromLocal) => {
    const src = pathInfo.source;
    const dest = pathInfo.destination;

    let statsSrc = null;
    if (fromLocal) {
        statsSrc = fs.statSync(CWD + src);

    }
    
    if (dest.substr(-5) === '.html') { // convert one file with a different name.
        if (statsSrc && !statsSrc.isFile()) {
            console.error('A directory can not be converted to a file');
        }
        return pathInfo.destination;
    }

    if (!fromLocal || statsSrc.isFile()) { //convert one file to a directory
        return dest + path.basename(src).replace('.md', '.html');
    }
    return ''; // save all files in src directory to destination directory. 
}
const isLocalPath = (arg) => {
    return arg.indexOf('http') === -1;
}
const generateMetaData = (title, tabs) => {
    return`<!--
topicTitle: ${title}
topicIds: [${tabs.join(', ')}]
-->`;
}
const generateHTMLFromMD = () => {
    pathData.forEach(({
        source,
        destination,
        topicTitle,
        topicIds
    }) => {
        const isLocal = isLocalPath(source);
        const dest = getDestination({source, destination}, isLocal);
        if (isLocal) {
            if (dest) { // convert one file
                convertLocalFile(source, dest);
            } else { //convert all files in the directory
                convertLocalFiles({source, destination});
            }
        } else {
            axios.get(source).then((response) => {
                generateMetaData(topicTitle, topicIds)
                const metaData = topicTitle && topicIds ? generateMetaData(topicTitle, topicIds) : '';
                saveFile(CWD + dest, metaData + convertFile(response.data));
            });
        }
    });
};

generateHTMLFromMD();