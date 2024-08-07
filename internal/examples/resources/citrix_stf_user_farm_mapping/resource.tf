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
            primary_farms = [ citrix_stf_store_farm.example-primary-store-farm.farm_name ]
            backup_farms = [ citrix_stf_store_farm.example-backup-store-farm.farm_name ]
            load_balance_mode = "LoadBalanced"
            farms_are_identical = true
        },
        {
            name = "EU2",
            aggregation_group_name = "EU2Users"
            primary_farms = [ citrix_stf_store_farm.example-secondary-store-farm.farm_name ]
            load_balance_mode = "Failover"
            farms_are_identical = false
        }
    ]

    // Add depends_on attribute to ensure the User Farm Mapping is created after the Store Service and Store Farms
    depends_on = [
        citrix_stf_store_service.example-stf-store-service,
        citrix_stf_store_farm.example-primary-store-farm,
        citrix_stf_store_farm.example-secondary-store-farm,
        citrix_stf_store_farm.example-backup-store-farm
    ]
}
