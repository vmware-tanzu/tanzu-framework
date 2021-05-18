// This utility will capture ALL current OS environmental variables
// and save them in .os-env-vars
const fs = require('fs');

const filename = '.os-env-vars';
const vars = JSON.stringify(process.env, null, 4);

try {
    fs.writeFile(filename, vars, 'utf8', function (err) {
        if (err) {
            return console.log(err);
        }
    });
} catch (e) {
    console.log(e);
}
