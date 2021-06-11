export interface CliFields {
    configPath: string;
    clusterType: string;
}
export class CliGenerator {
    getCli({
        configPath,
        clusterType
    }) {
        const clusterPrefix = (clusterType) ? clusterType : 'management';
        let command = `tanzu ${clusterPrefix}-cluster create`;
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
