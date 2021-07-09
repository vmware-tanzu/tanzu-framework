import { Observable, ReplaySubject } from 'rxjs';
import Broker from './broker';

import { TkgEventType } from './Messenger';

export abstract class WizardFormBase {

    private dataState = {};
    private errorState = {};

    constructor(private dataSources, private errorSpec) {
        this.dataSources.forEach(source => {
            // Initialize all data states with an empty array
            this.dataState[source] = new ReplaySubject(1);

            // Initialize all error states with a null
            this.errorState[source] = new ReplaySubject(1);

            // Reload data from backend upon notified
            Broker.messenger.getSubject(source)
                .subscribe((event) => this.retrieveData(source, event.payload));
        });
    }

    /**
     * Retrieve data for source
     * @param source TkgEventType
     */
    abstract retrieveDataForSource(source: TkgEventType, payload?: any): Observable<any>;

    /**
     * Generic method for retrieving data via apiClient.
     * @param source TkgEventType
     */
    private retrieveData(source: TkgEventType, payload?: any) {
        this.retrieveDataForSource(source, payload)
            .subscribe(
                (data => {
                    this.publishData(source, data);
                    this.publishError(source, null);
                }),
                (err => {
                    const error = err.error.message || err.message || JSON.stringify(err);
                    this.publishError(source, error);
                    this.publishData(source, []);   // Fixme: change '[]' to 'null' to make it more generic.
                })
            );
    }

    /**
     * Publish data for consumers to absorb
     * @param source TkgEventType
     * @param data the data to be published
     */
    publishData(source: TkgEventType, data: any) {
        this.dataState[source].next(data);
    }

    /**
     * Return the data stream for source
     * @param source TkgEventType
     */
    getDataStream(source: TkgEventType): Observable<any> {
        return this.dataState[source];
    }

    /**
     * Publish an error message for consumers to absorb.
     * @param source TkgEventType
     * @param error error message to publish
     */
    publishError(source: TkgEventType, error: string) {
        const message = error ? this.errorSpec[source] + " " + error : null;
        this.errorState[source].next(message);
    }

    /**
     * Return error stream for source
     * @param source TkgEventType
     */
    getErrorStream(source: TkgEventType): Observable<string> {
        return this.errorState[source];
    }

    /**
     * Converts ES6 map to stringifyable object
     * @param strMap ES6 map that will be converted
     */
    strMapToObj(strMap: Map<string, string>): { [key: string]: string; } {
        const obj = Object.create(null);
        for (const [k, v] of strMap) {
          obj[k] = v;
        }
        return obj;
    }
}
