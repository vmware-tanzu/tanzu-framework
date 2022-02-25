import { Messenger } from './Messenger';
import { AppDataService } from "./app-data.service";
import DataServiceRegistrar from './data-service-registrar';
import { UserDataService } from './user-data.service';
import { FieldMapUtilities } from '../../views/landing/wizard/shared/field-mapping/FieldMapUtilities';
import { ValidationService } from '../../views/landing/wizard/shared/validation/validation.service';
import { UserDataFormService } from './user-data-form.service';

/**
 *@Class AppServices - exports classes as static members to avoid injecting instances of these classes through a constructor
 */
export default class AppServices {
    static appDataService = new AppDataService();
    static dataServiceRegistrar = new DataServiceRegistrar();
    static fieldMapUtilities = new FieldMapUtilities(new ValidationService());
    static messenger = new Messenger();
    static userDataService = new UserDataService();
    static userDataFormService = new UserDataFormService();
}
