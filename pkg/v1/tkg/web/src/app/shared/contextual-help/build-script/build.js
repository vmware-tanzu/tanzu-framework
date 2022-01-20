const fs = require('fs');
const fm = require('html-frontmatter');
const removeHtmlComments = require('remove-html-comments');
const elasticlunr = require('elasticlunr');

const docsPath = 'src/contextualHelpDocs';

const lunrIndex = elasticlunr(function() {
    this.addField('topicIds');
    this.setRef('topicTitle');
    this.saveDocument(true);
});

const loadDoc = (filePath) => {
    const fileContent = fs.readFileSync(filePath, 'utf-8');
    const metaData = fm(fileContent);
    const htmlContent = removeHtmlComments(fileContent);
    lunrIndex.addDoc({
        ...metaData,
        htmlContent: htmlContent.data
    });
}

const getAllFiles = (dir) => {
    return new Promise((resolve, reject) => {
        const allPromises = []; 
        fs.readdir(dir, (err, files) => {
            if (err) {
                console.log(err);
                return;
            }
            files.forEach(file => {
                const filePath = `${dir}/${file}`;
                if (fs.statSync(filePath).isDirectory()) {
                    allPromises.push(getAllFiles(filePath))
                } else {
                    if (file.endsWith('html')) {
                        loadDoc(filePath);
                    }
                    allPromises.push(Promise.resolve());
                }
            });
            Promise.all(allPromises).then(resolve);
        });
    });
}

getAllFiles(docsPath).then(() => {
    fs.writeFileSync(docsPath + '/index.json', JSON.stringify(lunrIndex.toJSON()), {
        encoding: 'utf-8'
    });
    console.log('Elastic lunr index created successfully');
});


