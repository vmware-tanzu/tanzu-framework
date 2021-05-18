const puppeteer = require('puppeteer');
const fse = require('fs-extra');

async function run() {
    const browser = await puppeteer.launch();
    const page = await browser.newPage();

    const filePath = `${process.env.PWD}/index.html`;

    await page.goto(`file:///${process.env.PWD}/reports/index.html`, { waitUntil: 'networkidle0' });

    const html = await page.content();
    await fse.outputFile(filePath, html);

    try {
        // Overwrite the existing html with the prerendered one
        fse.moveSync(filePath, `${process.env.PWD}/reports/index.html`, { overwrite: true });
        fse.copySync("./src/report.css", `${process.env.PWD}/reports/report.css`, { overwrite: true });
        fse.copySync("./src/report.js", `${process.env.PWD}/reports/report.js`, { overwrite: true });
    } catch (e) {
        console.log(e);
    }

    browser.close();
}

run();