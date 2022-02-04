import { ModuleWithProviders } from '@angular/core';
import { HttpHeaders, HttpParams } from '@angular/common/http';
import * as i0 from "@angular/core";
export { APIClient } from './api-client.service';
export { APIClientInterface } from './api-client.interface';
export { GuardedAPIClient } from './guarded-api-client.service';
/**
 * provided options, headers and params will be used as default for each request
 */
export interface DefaultHttpOptions {
    headers?: {
        [key: string]: string;
    };
    params?: {
        [key: string]: string;
    };
    reportProgress?: boolean;
    withCredentials?: boolean;
}
export interface HttpOptions {
    headers?: HttpHeaders;
    params?: HttpParams;
    reportProgress?: boolean;
    withCredentials?: boolean;
}
export interface APIClientModuleConfig {
    domain?: string;
    guardResponses?: boolean;
    httpOptions?: DefaultHttpOptions;
}
export declare class APIClientModule {
    /**
     * Use this method in your root module to provide the APIClientModule
     *
     * If you are not providing
     * @param { APIClientModuleConfig } config
     * @returns { ModuleWithProviders }
     */
    static forRoot(config?: APIClientModuleConfig): ModuleWithProviders<APIClientModule>;
    static ɵfac: i0.ɵɵFactoryDeclaration<APIClientModule, never>;
    static ɵmod: i0.ɵɵNgModuleDeclaration<APIClientModule, never, never, never>;
    static ɵinj: i0.ɵɵInjectorDeclaration<APIClientModule>;
}
