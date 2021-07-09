// Angular imports
import { Component, Inject, PLATFORM_ID } from '@angular/core';
import { DOCUMENT, isPlatformBrowser } from '@angular/common';

/**
 * @class ThemeToggleComponent
 * ThemeToggleComponent allows user to switch between Clarity light and dark mode themes.
 */
@Component({
    selector: 'tkg-kickstart-ui-theme-toggle',
    templateUrl: './theme-toggle.component.html',
    styleUrls: ['./theme-toggle.component.scss']
})
export class ThemeToggleComponent {
    linkRef: HTMLLinkElement;

    themes = [
        { name: 'light', href: 'assets/css/clr-ui.min.css' },
        { name: 'dark', href: 'assets/css/clr-ui-dark.min.css' }
    ];

    darkBodyClass: string = 'dark';
    theme = this.themes[0];
    constructor(@Inject(DOCUMENT) private document: Document,
                @Inject(PLATFORM_ID) private platformId: Object) {
        if (isPlatformBrowser(this.platformId)) {
            try {
                const stored = localStorage.getItem('clr-theme');
                if (stored) {
                    this.theme = JSON.parse(stored);
                }
            } catch (err) {
                console.log(`Error retrieving clr-theme from local storage: ${err}`);
            }
            this.linkRef = this.document.createElement('link');
            this.linkRef.rel = 'stylesheet';
            this.linkRef.href = this.theme.href;
            this.document.querySelector('head').appendChild(this.linkRef);
            if (this.theme.name) {
                this.document.body.className = this.theme.name;
            }
        }
    }

    switchTheme() {
        if (this.theme.name === 'light') {
            this.theme = this.themes[1];
            this.document.body.className = this.darkBodyClass;
        } else {
            this.theme = this.themes[0];
            this.document.body.className = '';
        }
        localStorage.setItem('clr-theme', JSON.stringify(this.theme));
        this.linkRef.href = this.theme.href;
    }
}
