export interface CliFields {
    configPath: string;
    clusterType: string;
    clusterName: string;
}
export class CliGenerator {
    getCli({
        configPath,
        clusterType,
        clusterName
    }) {
        const clusterPrefix = (clusterType) ? clusterType : 'management';
        const clusterNameArg = (clusterName) ? ` ${clusterName} ` : '';
        let command = `tanzu ${clusterPrefix}-cluster create${clusterNameArg}`;
        const optionsMapping = [
            ['--file', configPath],
            ['-v', 6]
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
        return command;
    }
}
