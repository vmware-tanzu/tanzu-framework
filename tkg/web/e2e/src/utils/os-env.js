/**
 * Initialize process.env with setting from both .os-env-vars and .env
 * Settings in .env will be overwritten by the corresponding ones from
 * .os-env-vars silently.
 */

const fs = require('fs');
const dotenv = require("dotenv");

const OS_ENV_VARS_FILE = ".os-env-vars";

exports.config = () => {
    dotenv.config();

    try {
        if (fs.existsSync(OS_ENV_VARS_FILE)) {
            fs.readFile(OS_ENV_VARS_FILE, 'utf8', (err, data) => {
                if (err) {
                    console.log(err);
                } else {
                    if (data && data.length > 0) {
                        const vars = JSON.parse(data);
                        process.env = { ...process.env, ...vars };
                    }
                }
            })
        }
    } catch (e) {
        console.log(e);
    }
}

