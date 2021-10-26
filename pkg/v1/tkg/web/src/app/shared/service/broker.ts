import { Messenger } from './Messenger';
import { AppDataService } from "./app-data.service";

export default class Broker {
    static messenger = new Messenger();
    static appDataService = new AppDataService();
}
