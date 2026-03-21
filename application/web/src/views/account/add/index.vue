<template>
  <div class="app-container">
    <el-form ref="ruleForm" v-loading="loading" :model="ruleForm" :rules="rules" label-width="100px" class="account-form">
      <el-form-item label="姓名/名称" prop="account_name">
        <el-input v-model="ruleForm.account_name" />
      </el-form-item>

      <el-form-item label="角色类型" prop="role">
        <el-select v-model="ruleForm.role" placeholder="请选择角色类型" @change="onRoleChange">
          <el-option label="医生" value="doctor" />
          <el-option label="患者" value="patient" />
          <el-option label="药店" value="drugstore" />
          <el-option label="保险机构" value="insurance" />
        </el-select>
      </el-form-item>

      <template v-if="ruleForm.role === 'doctor'">
        <el-form-item label="所属医院" prop="hospital_name">
          <el-select v-model="ruleForm.hospital_name" placeholder="请选择所属医院">
            <el-option v-for="item in hospitals" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="所属科室" prop="department">
          <el-select v-model="ruleForm.department" placeholder="请选择所属科室">
            <el-option v-for="item in departments" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="职位" prop="title">
          <el-select v-model="ruleForm.title" placeholder="请选择职位">
            <el-option v-for="item in titles" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="性别" prop="gender">
          <el-select v-model="ruleForm.gender" placeholder="请选择性别">
            <el-option label="男" value="男" />
            <el-option label="女" value="女" />
          </el-select>
        </el-form-item>
        <el-form-item label="工号" prop="employee_no">
          <el-input v-model="ruleForm.employee_no" />
        </el-form-item>
      </template>

      <template v-if="ruleForm.role === 'patient'">
        <el-form-item label="身份证号" prop="id_card_no">
          <el-input v-model="ruleForm.id_card_no" />
        </el-form-item>
        <el-form-item label="医保卡号" prop="insurance_card_no">
          <el-input v-model="ruleForm.insurance_card_no" />
        </el-form-item>
        <el-form-item label="性别" prop="gender">
          <el-select v-model="ruleForm.gender" placeholder="请选择性别">
            <el-option label="男" value="男" />
            <el-option label="女" value="女" />
          </el-select>
        </el-form-item>
        <el-form-item label="年龄" prop="age">
          <el-input-number v-model="ruleForm.age" :min="0" :max="150" />
        </el-form-item>
        <el-form-item label="出生年月" prop="birth_date">
          <el-date-picker
            v-model="ruleForm.birth_date"
            type="date"
            placeholder="选择日期"
            value-format="yyyy-MM-dd"
          />
        </el-form-item>
        <el-form-item label="联系方式" prop="phone">
          <el-input v-model="ruleForm.phone" />
        </el-form-item>
      </template>

      <template v-if="ruleForm.role === 'drugstore'">
        <el-form-item label="所属医院" prop="hospital_name">
          <el-select v-model="ruleForm.hospital_name" placeholder="请选择所属医院">
            <el-option v-for="item in hospitals" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
      </template>

      <el-form-item>
        <el-button type="primary" @click="submitForm('ruleForm')">立即创建</el-button>
        <el-button @click="resetForm('ruleForm')">重置</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import { createAccount } from '@/api/accountV2'

export default {
  name: 'AddAccount',
  data() {
    return {
      hospitals: ['北京协和医院', '华西医院', '上海交通大学医学院附属瑞金医院', '广东省人民医院'],
      departments: ['内科', '外科', '骨科', '神经科', '儿科'],
      titles: ['住院医师', '主治医师', '副主任医师', '主任医师'],
      ruleForm: {
        account_name: '',
        role: '',
        hospital_id: '',
        hospital_name: '',
        department: '',
        title: '',
        gender: '',
        employee_no: '',
        id_card_no: '',
        insurance_card_no: '',
        age: null,
        birth_date: '',
        phone: ''
      },
      loading: false,
      rules: {
        account_name: [{ required: true, message: '请输入姓名/名称', trigger: 'blur' }],
        role: [{ required: true, message: '请选择角色类型', trigger: 'change' }],
        hospital_name: [{ validator: (rule, value, callback) => this.validateHospital(value, callback), trigger: 'change' }],
        department: [{ validator: (rule, value, callback) => this.validateDoctorField('department', value, callback), trigger: 'change' }],
        title: [{ validator: (rule, value, callback) => this.validateDoctorField('title', value, callback), trigger: 'change' }],
        gender: [{ validator: (rule, value, callback) => this.validateGender(value, callback), trigger: 'change' }],
        employee_no: [{ validator: (rule, value, callback) => this.validateDoctorField('employee_no', value, callback), trigger: 'blur' }],
        id_card_no: [{ validator: (rule, value, callback) => this.validatePatientField('id_card_no', value, callback), trigger: 'blur' }],
        insurance_card_no: [{ validator: (rule, value, callback) => this.validatePatientField('insurance_card_no', value, callback), trigger: 'blur' }],
        age: [{ validator: (rule, value, callback) => this.validatePatientField('age', value, callback), trigger: 'change' }],
        birth_date: [{ validator: (rule, value, callback) => this.validatePatientField('birth_date', value, callback), trigger: 'change' }],
        phone: [{ validator: (rule, value, callback) => this.validatePatientField('phone', value, callback), trigger: 'blur' }]
      }
    }
  },
  computed: {
    ...mapGetters([
      'account_id'
    ])
  },
  methods: {
    onRoleChange() {
      this.ruleForm.hospital_name = ''
      this.ruleForm.department = ''
      this.ruleForm.title = ''
      this.ruleForm.gender = ''
      this.ruleForm.employee_no = ''
      this.ruleForm.id_card_no = ''
      this.ruleForm.insurance_card_no = ''
      this.ruleForm.age = null
      this.ruleForm.birth_date = ''
      this.ruleForm.phone = ''
    },
    validateHospital(value, callback) {
      if ((this.ruleForm.role === 'doctor' || this.ruleForm.role === 'drugstore') && !value) {
        callback(new Error('请选择所属医院'))
        return
      }
      callback()
    },
    validateDoctorField(field, value, callback) {
      if (this.ruleForm.role === 'doctor' && !value) {
        const map = {
          department: '请选择所属科室',
          title: '请选择职位',
          employee_no: '请输入工号'
        }
        callback(new Error(map[field]))
        return
      }
      callback()
    },
    validatePatientField(field, value, callback) {
      if (this.ruleForm.role === 'patient' && (value === '' || value === null || value === undefined)) {
        const map = {
          id_card_no: '请输入身份证号',
          insurance_card_no: '请输入医保卡号',
          age: '请输入年龄',
          birth_date: '请选择出生年月',
          phone: '请输入联系方式'
        }
        callback(new Error(map[field]))
        return
      }
      callback()
    },
    validateGender(value, callback) {
      if ((this.ruleForm.role === 'doctor' || this.ruleForm.role === 'patient') && !value) {
        callback(new Error('请选择性别'))
        return
      }
      callback()
    },
    submitForm(formName) {
      this.$refs[formName].validate((valid) => {
        if (!valid) return false

        this.$confirm('是否立即创建?', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'success'
        }).then(() => {
          this.loading = true
          createAccount({
            operator: this.account_id,
            account_name: this.ruleForm.account_name,
            role: this.ruleForm.role,
            hospital_id: this.ruleForm.hospital_id,
            hospital_name: this.ruleForm.hospital_name,
            department: this.ruleForm.department,
            title: this.ruleForm.title,
            gender: this.ruleForm.gender,
            employee_no: this.ruleForm.employee_no,
            id_card_no: this.ruleForm.id_card_no,
            insurance_card_no: this.ruleForm.insurance_card_no,
            age: this.ruleForm.age === null ? '' : String(this.ruleForm.age),
            birth_date: this.ruleForm.birth_date,
            phone: this.ruleForm.phone
          }).then(response => {
            this.loading = false
            if (response !== null) {
              this.$message({ type: 'success', message: '创建成功!' })
              this.resetForm(formName)
            } else {
              this.$message({ type: 'error', message: '创建失败!' })
            }
          }).catch(() => {
            this.loading = false
          })
        }).catch(() => {
          this.loading = false
          this.$message({ type: 'info', message: '已取消创建' })
        })
      })
    },
    resetForm(formName) {
      this.$refs[formName].resetFields()
      this.onRoleChange()
    }
  }
}
</script>

<style scoped>
.account-form {
  max-width: 680px;
}

.account-form .el-select,
.account-form .el-input,
.account-form .el-date-editor {
  width: 100%;
}
</style>
