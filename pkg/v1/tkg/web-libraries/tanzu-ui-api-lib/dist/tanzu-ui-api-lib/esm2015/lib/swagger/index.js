/* tslint:disable */
import { NgModule } from '@angular/core';
import { APIClient, USE_DOMAIN, USE_HTTP_OPTIONS } from './api-client.service';
import { GuardedAPIClient } from './guarded-api-client.service';
import * as i0 from "@angular/core";
export { APIClient } from './api-client.service';
export { GuardedAPIClient } from './guarded-api-client.service';
export class APIClientModule {
    /**
     * Use this method in your root module to provide the APIClientModule
     *
     * If you are not providing
     * @param { APIClientModuleConfig } config
     * @returns { ModuleWithProviders }
     */
    static forRoot(config = {}) {
        return {
            ngModule: APIClientModule,
            providers: [
                ...(config.domain != null ? [{ provide: USE_DOMAIN, useValue: config.domain }] : []),
                ...(config.httpOptions ? [{ provide: USE_HTTP_OPTIONS, useValue: config.httpOptions }] : []),
                ...(config.guardResponses ? [{ provide: APIClient, useClass: GuardedAPIClient }] : [APIClient]),
            ]
        };
    }
}
APIClientModule.ɵfac = i0.ɵɵngDeclareFactory({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0, type: APIClientModule, deps: [], target: i0.ɵɵFactoryTarget.NgModule });
APIClientModule.ɵmod = i0.ɵɵngDeclareNgModule({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0, type: APIClientModule });
APIClientModule.ɵinj = i0.ɵɵngDeclareInjector({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0, type: APIClientModule });
i0.ɵɵngDeclareClassMetadata({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0, type: APIClientModule, decorators: [{
            type: NgModule,
            args: [{}]
        }] });
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiaW5kZXguanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyIuLi8uLi8uLi8uLi8uLi9wcm9qZWN0cy90YW56dS11aS1hcGktbGliL3NyYy9saWIvc3dhZ2dlci9pbmRleC50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQSxvQkFBb0I7QUFFcEIsT0FBTyxFQUFFLFFBQVEsRUFBdUIsTUFBTSxlQUFlLENBQUM7QUFFOUQsT0FBTyxFQUFFLFNBQVMsRUFBRSxVQUFVLEVBQUUsZ0JBQWdCLEVBQUUsTUFBTSxzQkFBc0IsQ0FBQztBQUMvRSxPQUFPLEVBQUUsZ0JBQWdCLEVBQUUsTUFBTSw4QkFBOEIsQ0FBQzs7QUFFaEUsT0FBTyxFQUFFLFNBQVMsRUFBRSxNQUFNLHNCQUFzQixDQUFDO0FBRWpELE9BQU8sRUFBRSxnQkFBZ0IsRUFBRSxNQUFNLDhCQUE4QixDQUFDO0FBMEJoRSxNQUFNLE9BQU8sZUFBZTtJQUMxQjs7Ozs7O09BTUc7SUFDSCxNQUFNLENBQUMsT0FBTyxDQUFDLFNBQWdDLEVBQUU7UUFDL0MsT0FBTztZQUNMLFFBQVEsRUFBRSxlQUFlO1lBQ3pCLFNBQVMsRUFBRTtnQkFDVCxHQUFHLENBQUMsTUFBTSxDQUFDLE1BQU0sSUFBSSxJQUFJLENBQUMsQ0FBQyxDQUFDLENBQUMsRUFBQyxPQUFPLEVBQUUsVUFBVSxFQUFFLFFBQVEsRUFBRSxNQUFNLENBQUMsTUFBTSxFQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsRUFBRSxDQUFDO2dCQUNsRixHQUFHLENBQUMsTUFBTSxDQUFDLFdBQVcsQ0FBQyxDQUFDLENBQUMsQ0FBQyxFQUFDLE9BQU8sRUFBRSxnQkFBZ0IsRUFBRSxRQUFRLEVBQUUsTUFBTSxDQUFDLFdBQVcsRUFBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLEVBQUUsQ0FBQztnQkFDMUYsR0FBRyxDQUFDLE1BQU0sQ0FBQyxjQUFjLENBQUMsQ0FBQyxDQUFDLENBQUMsRUFBQyxPQUFPLEVBQUUsU0FBUyxFQUFFLFFBQVEsRUFBRSxnQkFBZ0IsRUFBRSxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsU0FBUyxDQUFDLENBQUM7YUFDL0Y7U0FDRixDQUFDO0lBQ0osQ0FBQzs7NkdBakJVLGVBQWU7OEdBQWYsZUFBZTs4R0FBZixlQUFlOzRGQUFmLGVBQWU7a0JBRDNCLFFBQVE7bUJBQUMsRUFBRSIsInNvdXJjZXNDb250ZW50IjpbIi8qIHRzbGludDpkaXNhYmxlICovXG5cbmltcG9ydCB7IE5nTW9kdWxlLCBNb2R1bGVXaXRoUHJvdmlkZXJzIH0gZnJvbSAnQGFuZ3VsYXIvY29yZSc7XG5pbXBvcnQgeyBIdHRwSGVhZGVycywgSHR0cFBhcmFtcyB9IGZyb20gJ0Bhbmd1bGFyL2NvbW1vbi9odHRwJztcbmltcG9ydCB7IEFQSUNsaWVudCwgVVNFX0RPTUFJTiwgVVNFX0hUVFBfT1BUSU9OUyB9IGZyb20gJy4vYXBpLWNsaWVudC5zZXJ2aWNlJztcbmltcG9ydCB7IEd1YXJkZWRBUElDbGllbnQgfSBmcm9tICcuL2d1YXJkZWQtYXBpLWNsaWVudC5zZXJ2aWNlJztcblxuZXhwb3J0IHsgQVBJQ2xpZW50IH0gZnJvbSAnLi9hcGktY2xpZW50LnNlcnZpY2UnO1xuZXhwb3J0IHsgQVBJQ2xpZW50SW50ZXJmYWNlIH0gZnJvbSAnLi9hcGktY2xpZW50LmludGVyZmFjZSc7XG5leHBvcnQgeyBHdWFyZGVkQVBJQ2xpZW50IH0gZnJvbSAnLi9ndWFyZGVkLWFwaS1jbGllbnQuc2VydmljZSc7XG5cbi8qKlxuICogcHJvdmlkZWQgb3B0aW9ucywgaGVhZGVycyBhbmQgcGFyYW1zIHdpbGwgYmUgdXNlZCBhcyBkZWZhdWx0IGZvciBlYWNoIHJlcXVlc3RcbiAqL1xuZXhwb3J0IGludGVyZmFjZSBEZWZhdWx0SHR0cE9wdGlvbnMge1xuICBoZWFkZXJzPzoge1trZXk6IHN0cmluZ106IHN0cmluZ307XG4gIHBhcmFtcz86IHtba2V5OiBzdHJpbmddOiBzdHJpbmd9O1xuICByZXBvcnRQcm9ncmVzcz86IGJvb2xlYW47XG4gIHdpdGhDcmVkZW50aWFscz86IGJvb2xlYW47XG59XG5cbmV4cG9ydCBpbnRlcmZhY2UgSHR0cE9wdGlvbnMge1xuICBoZWFkZXJzPzogSHR0cEhlYWRlcnM7XG4gIHBhcmFtcz86IEh0dHBQYXJhbXM7XG4gIHJlcG9ydFByb2dyZXNzPzogYm9vbGVhbjtcbiAgd2l0aENyZWRlbnRpYWxzPzogYm9vbGVhbjtcbn1cblxuZXhwb3J0IGludGVyZmFjZSBBUElDbGllbnRNb2R1bGVDb25maWcge1xuICBkb21haW4/OiBzdHJpbmc7XG4gIGd1YXJkUmVzcG9uc2VzPzogYm9vbGVhbjsgLy8gdmFsaWRhdGUgcmVzcG9uc2VzIHdpdGggdHlwZSBndWFyZHNcbiAgaHR0cE9wdGlvbnM/OiBEZWZhdWx0SHR0cE9wdGlvbnM7XG59XG5cbkBOZ01vZHVsZSh7fSlcbmV4cG9ydCBjbGFzcyBBUElDbGllbnRNb2R1bGUge1xuICAvKipcbiAgICogVXNlIHRoaXMgbWV0aG9kIGluIHlvdXIgcm9vdCBtb2R1bGUgdG8gcHJvdmlkZSB0aGUgQVBJQ2xpZW50TW9kdWxlXG4gICAqXG4gICAqIElmIHlvdSBhcmUgbm90IHByb3ZpZGluZ1xuICAgKiBAcGFyYW0geyBBUElDbGllbnRNb2R1bGVDb25maWcgfSBjb25maWdcbiAgICogQHJldHVybnMgeyBNb2R1bGVXaXRoUHJvdmlkZXJzIH1cbiAgICovXG4gIHN0YXRpYyBmb3JSb290KGNvbmZpZzogQVBJQ2xpZW50TW9kdWxlQ29uZmlnID0ge30pOiBNb2R1bGVXaXRoUHJvdmlkZXJzPEFQSUNsaWVudE1vZHVsZT4ge1xuICAgIHJldHVybiB7XG4gICAgICBuZ01vZHVsZTogQVBJQ2xpZW50TW9kdWxlLFxuICAgICAgcHJvdmlkZXJzOiBbXG4gICAgICAgIC4uLihjb25maWcuZG9tYWluICE9IG51bGwgPyBbe3Byb3ZpZGU6IFVTRV9ET01BSU4sIHVzZVZhbHVlOiBjb25maWcuZG9tYWlufV0gOiBbXSksXG4gICAgICAgIC4uLihjb25maWcuaHR0cE9wdGlvbnMgPyBbe3Byb3ZpZGU6IFVTRV9IVFRQX09QVElPTlMsIHVzZVZhbHVlOiBjb25maWcuaHR0cE9wdGlvbnN9XSA6IFtdKSxcbiAgICAgICAgLi4uKGNvbmZpZy5ndWFyZFJlc3BvbnNlcyA/IFt7cHJvdmlkZTogQVBJQ2xpZW50LCB1c2VDbGFzczogR3VhcmRlZEFQSUNsaWVudCB9XSA6IFtBUElDbGllbnRdKSxcbiAgICAgIF1cbiAgICB9O1xuICB9XG59XG4iXX0=