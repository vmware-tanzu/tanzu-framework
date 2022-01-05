import { Messenger } from './Messenger';
import { AppDataService } from "./app-data.service";
import { UserDataService } from './user-data.service';
import { FieldMapUtilities } from '../../views/landing/wizard/shared/field-mapping/FieldMapUtilities';
import { ValidationService } from '../../views/landing/wizard/shared/validation/validation.service';

/**
 *@Class Broker - exports classes as static members to avoid injecting instances of these classes through a constructor

 Usage - AppServices.messenger.methodName() - access methods from static class without need for a Messenger instance
 Usage - AppServices.appDataService.methodName() - access methods from static class without need for an AppDataService instance
 */
export default class Broker {
    static appDataService = new AppDataService();
    static fieldMapUtilities = new FieldMapUtilities(new ValidationService());
    static messenger = new Messenger();
    static userDataService = new UserDataService();
}
