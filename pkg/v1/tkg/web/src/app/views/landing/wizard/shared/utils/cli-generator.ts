import { ClusterType } from '../constants/wizard.constants';

export interface CliFields {
    configPath: string;
    clusterType: ClusterType;
    clusterName: string;
    extendCliCmds: Array<{isPrefixOfCreateCmd: boolean, cmdStr: string}>;
}
export class CliGenerator {
    getCli(cliFields: CliFields) {
        const clusterType = (cliFields.clusterType) ? cliFields.clusterType : ClusterType.Management;
        const clusterPrefix = '' + clusterType;
        const clusterNameArg = (cliFields.clusterName) ? ` ${cliFields.clusterName} ` : '';
        let command = `tanzu ${clusterPrefix}-cluster create${clusterNameArg}`;
        const optionsMapping = [
            ['--file', cliFields.configPath],
            ['-v', '6']
        ];
        optionsMapping.forEach(option => {
            if (option[1] || typeof option[1] === 'boolean') {
                try {
                    const url = new URL(option[1]);
                    command += ` ${option[0]} '${url.href}'`;
                } catch (error) {
                    command += ` ${option[0]} ${option[1]}`;
                }
            }
        })

        const extendCliCmdsArray: Array<{isPrefixOfCreateCmd: boolean, cmdStr: string}> = cliFields.extendCliCmds;
        extendCliCmdsArray.forEach(item => {
            if (item.isPrefixOfCreateCmd) {
                command = item.cmdStr + " && " + command;
            } else {
                command = command  + " && " + item.cmdStr;
            }
        })
        return command;
    }
}
