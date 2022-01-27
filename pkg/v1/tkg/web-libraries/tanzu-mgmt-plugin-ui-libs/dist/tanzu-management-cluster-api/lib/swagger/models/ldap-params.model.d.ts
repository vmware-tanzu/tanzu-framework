export interface LdapParams {
    ldap_bind_dn?: string;
    ldap_bind_password?: string;
    ldap_group_search_base_dn?: string;
    ldap_group_search_filter?: string;
    ldap_group_search_group_attr?: string;
    ldap_group_search_name_attr?: string;
    ldap_group_search_user_attr?: string;
    ldap_root_ca?: string;
    ldap_test_group?: string;
    ldap_test_user?: string;
    ldap_url?: string;
    ldap_user_search_base_dn?: string;
    ldap_user_search_email_attr?: string;
    ldap_user_search_filter?: string;
    ldap_user_search_id_attr?: string;
    ldap_user_search_name_attr?: string;
    ldap_user_search_username?: string;
}
