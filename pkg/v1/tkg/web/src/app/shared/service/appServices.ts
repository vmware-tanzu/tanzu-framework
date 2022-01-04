import { Messenger } from './Messenger';
import { AppDataService } from "./app-data.service";
import DataServiceRegistrar from './data-service-registrar';

/**
 *@Class AppServices - exports classes as static members to avoid injecting instances of these classes through a constructor

 Usage - AppServices.messenger.methodName() - access methods from static class without need for a Messenger instance
 Usage - AppServices.appDataService.methodName() - access methods from static class without need for an AppDataService instance
 */
export default class AppServices {
    static messenger = new Messenger();
    static appDataService = new AppDataService();
    static dataServiceRegistrar = new DataServiceRegistrar();
}
