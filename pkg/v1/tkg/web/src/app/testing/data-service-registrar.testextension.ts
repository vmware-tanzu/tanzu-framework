import DataServiceRegistrar from '../shared/service/data-service-registrar';
import { TkgEventType } from '../shared/service/Messenger';
import { Observable } from 'rxjs';

export class DataServiceRegistrarTestExtension extends DataServiceRegistrar {
    public hasEntry(eventType: TkgEventType): boolean {
        const eventEntry = this.getEntry(eventType);
        return eventEntry !== null && eventEntry !== undefined;
    }

    public simulateError(eventType: TkgEventType, errMsg: string) {
        if (!this.hasEntry(eventType)) {
            console.log('No event registration found for ' + eventType);
        }
        this.getEntry(eventType).errorStream.next(errMsg);
    }

    public simulateData(eventType: TkgEventType, data: any) {
        if (!this.hasEntry(eventType)) {
            console.log('No event registration found for ' + eventType);
        }
        this.getEntry(eventType).dataStream.next(data);
    }

    public simulateRegistration<OBJ>(eventType: TkgEventType) {
        this.register<OBJ>(eventType, () => new Observable<OBJ[]>());
    }
}
