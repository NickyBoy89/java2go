import unittest
# import os

# Import all the functionality from the main file
import java2go

class TestSimpleJavaFile(unittest.TestCase):

    def test_conversion(self):

        simpleJava = ''
        simpleGolang = ''

        with open('testsnippets/simple.java', 'r') as javaFile:
            simpleJava = javaFile.read()

        with open('testsnippets/simple.go', 'r') as goFile:
            simpleGolang = goFile.read()

        self.assertEqual(java2go.convert(simpleJava), simpleGolang)

if __name__ == '__main__':
    unittest.main()
