import { Messenger } from './Messenger';
import { AppDataService } from "./app-data.service";

/**
 *@Class Broker - exports classes as static members to avoid injecting instances of these classes through a constructor

 Usage - Broker.messenger.methodName() - access methods from static class without need for a Messenger instance
 Usage - Broker.appDataService.methodName() - access methods from static class without need for an AppDataService instance
 */
export default class Broker {
    static messenger = new Messenger();
    static appDataService = new AppDataService();
}
