// @ts-check
// Protractor configuration file, see link for more information
// https://github.com/angular/protractor/blob/master/lib/config.ts

const { browser } = require('protractor');
const fs = require('fs');

const failFast = require('protractor-fail-fast');
const { SpecReporter } = require('jasmine-spec-reporter');

const HtmlScreenshotReporter = require('protractor-jasmine2-screenshot-reporter');
require('./src/utils/os-env').config();

const dashboardReporter = new HtmlScreenshotReporter({
    dest: 'reports',
    filename: 'index.html',
    showSummary: true,
    captureOnlyFailedSpecs: false,      // capture all screenshots
    reportOnlyFailedSpecs: false,
    showQuickLinks: false,
    showConfiguration: false,
    reportTitle: "tkg-cli UI e2e Report",
    userJs: 'report.js',
    userCss: 'report.css'
});

/**
 * Creates marker file based on the test result so that
 * other scripts can act accordingly.
 */
const whistleblower = (function () {
    const whistle = '.whistle';
    return {
        createWhistle: function () {
            try {
                fs.closeSync(fs.openSync(whistle, 'w'));
            } catch (e) {
                console.log(e);
            }
        },
        removeWhistle: function () {
            try {
                if (fs.existsSync(whistle)) {
                    fs.unlinkSync(whistle);
                }
            } catch (e) {
                console.log(e);
            }
        }
    }
})();

const chromeOptions = process.env.CHROME_OPTIONS ?
    { args: ['window-size=1920,1200'] } :
    { args: ["--headless", 'window-size=1920,1200'] };

const targetSpecs = process.env.E2E_SPEC ? `./src/${process.env.E2E_SPEC}.e2e-spec.ts` : "./src/**/*.e2e-spec.ts";

/**
 * @type { import("protractor").Config }
 */
exports.config = {
    params: {
        ...process.env
    },
    plugins: [
        failFast.init(),
    ],
    allScriptsTimeout: 11000,
    specs: [
        targetSpecs
    ],
    capabilities: {
        chromeOptions,
        browserName: 'chrome',
        'shardTestFiles': true,
        'maxInstances': 5
    },
    directConnect: false,
    baseUrl: `http://${process.env.SERVER_URL}/`,
    seleniumAddress: 'http://localhost:4444/wd/hub',
    framework: 'jasmine',
    jasmineNodeOpts: {
        showColors: true,
        defaultTimeoutInterval: 6000000,
        print: function () { }
    },
    beforeLaunch: function () {
        return new Promise(function (resolve) {
            dashboardReporter.beforeLaunch(resolve);
        });
    },
    onPrepare() {
        whistleblower.removeWhistle();

        require('ts-node').register({
            project: require('path').join(__dirname, './tsconfig.json')
        });

        jasmine.getEnv().addReporter(dashboardReporter);
        jasmine.getEnv().addReporter(new SpecReporter(
            {
                spec: {
                    displayStacktrace: false
                },
                summary: {
                    displayErrorMessages: false,
                    displayStacktrace: false,
                    displaySuccessful: true,
                    displayFailed: true,
                    displayPending: true,
                    displayDuration: true
                },
                colors: {
                    successful: 'green',
                    failed: 'red',
                    pending: 'yellow'
                },
                prefixes: {
                    successful: '✓ ',
                    failed: '✗ ',
                    pending: '* '
                }
            }
        ));
    },
    onComplete: function () {
    },

    afterLaunch: function (exitCode) {
        if (exitCode) {
            whistleblower.createWhistle();
        }
        failFast.clean(); // Removes the fail file once all test runners have completed.
        return new Promise(function (resolve) {
            dashboardReporter.afterLaunch(resolve.bind(this, 0));
        });
    },
};