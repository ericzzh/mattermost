
export MM_SQLSETTINGS_DRIVERNAME=mysql
echo "**********ZZH MOD*********  Set MM_SQLSETTINGS_DRIVERNAME: $MM_SQLSETTINGS_DRIVERNAME" 
export MM_SQLSETTINGS_DATASOURCE="mmuser:mostest@tcp(localhost:3306)/mattermost_test?charset=utf8mb4,utf8\u0026readTimeout=30s\u0026writeTimeout=30s\u0026multiStatements=true"
echo "***********ZZH MOD*********  Set MM_SQLSETTINGS_DATASOURCE: $MM_SQLSETTINGS_DATASOURCE" 
export MM_SERVICESETTINGS_ENABLELOCALMODE=true
echo "***********ZZH MOD*********  Set MM_SERVICESETTINGS_ENABLELOCALMODE: $MM_SERVICESETTINGS_ENABLELOCALMODE" 
export ZZH_IF_OURS="true"
echo "***********ZZH MOD*********  Set ZZH_IF_OURS: $ZZH_IF_OURS" 
export ZZH_MOCK_LIC_PATH=~/go/src/mattermost-server/mocklicense/mocklicense.json
echo "***********ZZH MOD*********  Set ZZH_MOCK_LIC_PATH: $ZZH_MOCK_LIC_PATH" 
export MM_SERVER_PATH="/Users/zzh/go/src/mattermost-server"
echo "***********ZZH MOD*********  Set MM_SERVER_PATH: $MM_SERVER_PATH" 
