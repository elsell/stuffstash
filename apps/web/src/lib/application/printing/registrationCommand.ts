function shellArgument(value:string){return `'${value.replaceAll("'", "'\\''")}'`;}
export function printerRegistrationCommand(apiBaseUrl:string,computerName:string){
 return `./stuffstash --server ${shellArgument(apiBaseUrl)} connectors print register --name ${shellArgument(computerName.trim())}`;
}
