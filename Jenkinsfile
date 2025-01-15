def teamEmail = 'dbtech@demonware.net'
def teamSlackChannel = '#platform-dbtech-builds'
def gitVars = ''
Boolean slackNotificationSent = true
List<String> branchesToNotifyOn = ['v18.0.7-demonware']

/*
to get list of tests:
source build.env && go run test.go --dry-run --skip-build
*/

def test_set_dict = [
    'backup_tests' : [
        'backup',
        'backup_mysqlctld',
        'backup_pitr',
        'backup_pitr_xtrabackup',
        'backup_xtrabackup_xbstream',
        'backup_xtrabackup_xbstream_lz4'
    ],
    'encryption_tests' : [
        'encrypted_replication',
        'encrypted_transport'
    ],
    'onlineddl_tests' : [
        'onlineddl_ghost',
        'onlineddl_revert',
        'onlineddl_scheduler',
        'onlineddl_vrepl_stress_suite'
    ],
    'tabletgateway_tests' : [
        'tabletgateway_buffer_reparent',
        'tabletgateway_buffer_reshard'
    ],
    'tabletmanager_tests' : [
        'tabletmanager',
        'tabletmanager_replication_manager',
        'tabletmanager_tablegc',
        'tabletmanager_throttler_topo',
        'tabletmanager_zk2'
    ],
    'vindex_tests' : [
        'vindex_bindvars',
        'vindex_secondary'
    ],
    'vdiff' : [
        'vdiff2',
        'vdiff_multiple_movetables_test.go'
    ],
    'vreplication_tests' : [
        'vreplication_basic',
        'vreplication_cellalias',
        'vreplication_fk',
        'vreplication_movetables_buffering',
        'vreplication_movetables_tz',
        'vreplication_multicell',
        'vreplication_partial_movetables_basic',
        'vreplication_partial_movetables_sequences',
        'vreplication_vschema_load',
        'vreplication_vtctl_migrate',
        'vreplication_vtctld_migrate',
        'vreplication_vtctldclient_cli',
        'vreplication_vtctldclient_materialize'
    ],
    'vstream_tests' : [
        'vstream_failover',
        'vstream_flush_binlog',
        'vstream_stoponreshard_false',
        'vstream_stoponreshard_true',
        'vstream_with_keyspaces_to_watch'
    ],
    'vtgate_1_tests' : [
        'vtgate_concurrentdml',
        'vtgate_dbddlplugin',
        'vtgate_foreignkey',
        'vtgate_foreignkey_stress',
        'vtgate_godriver',
        'vtgate_grpc_api',
        'vtgate_kill',
        'vtgate_queries_aggregation',
        'vtgate_queries_aggregation_partial_keyspace',
        'vtgate_queries_foundrows',
        'vtgate_queries_foundrows_partial_keyspace',
        'vtgate_queries_informationschema',
        'vtgate_queries_informationschema_partial_keyspace',
        'vtgate_queries_insert',
        'vtgate_queries_insert_partial_keyspace',
        'vtgate_queries_normalize',
        'vtgate_queries_normalize_partial_keyspace',
        'vtgate_queries_orderby',
        'vtgate_queries_orderby_partial_keyspace',
        'vtgate_queries_random',
        'vtgate_queries_reference',
        'vtgate_queries_subquery',
        'vtgate_queries_subquery_partial_keyspace',
        'vtgate_queries_timeout',
        'vtgate_queries_tpch',
        'vtgate_queries_union',
        'vtgate_queries_union_partial_keyspace',
        'vtgate_queries_vexplain',
        'vtgate_readafterwrite',
        'vtgate_reserved_conn1',
        'vtgate_reserved_conn2',
        'vtgate_reserved_conn3',
        'vtgate_reserved_conn4'
    ],
    'vtgate_2_tests' : [
        'vtgate_schema',
        'vtgate_schematracker_loadkeyspace',
        'vtgate_schematracker_restarttablet',
        'vtgate_schematracker_sharded',
        'vtgate_schematracker_sharded_prs',
        'vtgate_schematracker_unsharded',
        'vtgate_sequence',
        'vtgate_sequence_partial_keyspace',
        'vtgate_setstatement',
        'vtgate_tablet_healthcheck_cache',
        'vtgate_transaction',
        'vtgate_transaction_partial_exec',
        'vtgate_transaction_partial_exec_partial_keyspace',
        'vtgate_transaction_partial_keyspace',
        'vtgate_transaction_restart',
        'vtgate_transaction_rollback',
        'vtgate_transaction_rollback_partial_keyspace',
        'vtgate_transaction_single',
        'vtgate_transaction_single_partial_keyspace',
        'vtgate_unsharded',
        'vtgate_vschema',
        'vtgate_watchkeyspace'
    ],
    'vtorc_tests' : [
        'vtorc_api',
        'vtorc_general',
        'vtorc_primary_failure',
        'vtorc_readtopologyinstance'
    ],
    'unit_tests' : [
        'unit'
    ],
    'other_tests' : [
        'cellalias',
        'client_test',
        'clustertest',
        'consolidator',
        'e2e_race',
        'emergencyreparent',
        'errs_as_warns',
        'errs_as_warns_partial_keyspace',
        'java',
        'keyspace',
        'local_example',
        'messaging',
        'mysql_server',
        'mysqlctl',
        'mysqlctld',
        'newfeatures',
        'plannedreparent',
        'prefixfanout',
        'prepare_statement',
        'prscomplex',
        'prssettingspool',
        'recovery',
        'region_example',
        'sharded',
        'sidecardb',
        'topo_connection_cache',
        'topo_consul',
        'topo_etcd2',
        'topo_zk2',
        'vault',
        'vttest_sample',
        'xb_recovery'
    ]
]

/*  Flakey or Failed everytime tests:
'Failed':[pitr, vtgate_queries_misc, vtgate_queries_derived, vreplication_mariadb_to_mysql,
vtgate, vreplication_across_db_versions, vtgate_gen4, onlineddl_vrepl_suite, e2e,
backup_xtrabackup, schemadiff_vrepl, vtgate_partial_keyspace, vreplication_migrate,onlineddl_vrepl_stress,
vtgate_mysql80, upgrade, vtgate_queries_misc_partial_keyspace, vreplication_v2, vreplication_materialize],
'Flakey':[unit_race, backup_only, tabletgateway, tabletmanager_consul, onlineddl_vrepl, vreplication_copy_parallel]
*/

pipeline {
    agent { label 'debian12' }

    options {
        timestamps()
        disableConcurrentBuilds()
    }
    stages {
        stage('Test') {
            matrix {
                agent { label 'debian12' }

                axes {
                    axis {
                        name 'test_set_name'
                        values 'backup_tests', 'encryption_tests', 'onlineddl_tests', 'tabletgateway_tests', 'tabletmanager_tests', 'vdiff', 'vindex_tests', 'vreplication_tests', 'vstream_tests', 'vtgate_1_tests', 'vtgate_2_tests', 'unit_tests', 'vtorc_tests', 'other_tests'
                    }
                }
                stages {
                    stage('prepare') {
                        steps {
                            echo "Prepare test environment for ${test_set_name}"
                            sh '''
                                sudo rm -rf /usr/local/go
                                sudo tar -C /usr/local -xzf ./dw-testing/go1.23.8.linux-amd64.tar.gz
                                . ./jenkins.env
                                go version
                            '''
                        }
                    } //  stage('prepare')
                    stage('run tests') {
                        steps {
                            script {
                                def test_set = test_set_dict[test_set_name]
                                // run tests
                                test_set.each { test ->
                                    catchError(buildResult: 'FAILURE', stageResult: 'FAILURE') {
                                        try {
                                            sh ". ./jenkins.env && go run test.go --flavor=percona80 ${test}"
                                        } catch (Exception e) {
                                            // Archive logs if test fails
                                            archiveArtifacts '_test/**'
                                            error("Build failed due to exception: ${e.message}")
                                        } // try - catch
                                    } // catchError
                                } // test_set.each
                            } // script
                        } //steps
                    } //stage('run tests')
                } // stages of matrix
            } // matrix
        } // stage('Test')
    } // stages
    post {
        failure {
            script {
                if (slackNotificationSent) {
                    slackBuildReport(teamSlackChannel)
                }
            }
        }
        changed {
            script {
                if (slackNotificationSent && currentBuild.currentResult == 'SUCCESS') {
                    slackBuildReport(teamSlackChannel)
                }
            }
        }

        cleanup {
            // https://github.ihs.demonware.net/Demonware/jenkins-shared-library/blob/master/vars/
            cleanupAgent()
        }
    } // post
} // pipeline