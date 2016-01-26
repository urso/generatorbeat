from generatorbeat import BaseTest

import os


class Test(BaseTest):

    def test_base(self):
        """
        Basic test with exiting Generatorbeat normally
        """
        self.render_config_template(
                path=os.path.abspath(self.working_dir) + "/log/*"
        )

        generatorbeat_proc = self.start_beat()
        self.wait_until( lambda: self.log_contains("generatorbeat is running"))
        exit_code = generatorbeat_proc.kill_and_wait()
        assert exit_code == 0
