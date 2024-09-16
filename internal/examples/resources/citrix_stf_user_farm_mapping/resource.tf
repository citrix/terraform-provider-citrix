resource "citrix_stf_user_farm_mapping" "example-stf-user-farm-mapping" {
    name = "Example STFUserFarmMapping"
    store_virtual_path = citrix_stf_store_service.example-stf-store-service.virtual_path
    group_members = [
        {
            group_name = "TestGroup1"
            account_sid = "{First Account Sid}"
        },
        {
            group_name = "TestGroup2"
            account_sid = "{Second Account Sid}"
        }
    ]
    equivalent_farm_sets = [
        {
            name = "EU1",
            aggregation_group_name = "EU1Users"
            primary_farms = [citrix_stf_store_service.example-stf-store-service.farms.0.farm_name]
            backup_farms = [citrix_stf_store_service.example-stf-store-service.farms.1.farm_name]
            load_balance_mode = "LoadBalanced"
            farms_are_identical = true
        },
        {
            name = "EU2",
            aggregation_group_name = "EU2Users"
            primary_farms = [citrix_stf_store_service.example-stf-store-service.farms.1.farm_name]
            load_balance_mode = "Failover"
            farms_are_identical = false
        }
    ]


}
